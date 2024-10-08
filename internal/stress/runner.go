package stress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"kasm-stress-test/internal/api"
	"kasm-stress-test/internal/config"
	"kasm-stress-test/internal/models"
	"kasm-stress-test/internal/utils"
)

type Runner struct {
	client         *api.Client
	config         *config.Config
	username       string
	sessionNum     utils.IntFlag
	command        string
	kasmsToDestroy []string
	UserID         string
	wg             sync.WaitGroup
	statusCallback func(sessionNumber int, status string, duration time.Duration)
}

func NewRunner(cfg *config.Config, username string, sessionNum utils.IntFlag, command string) *Runner {
	return &Runner{
		client:         api.NewClient(cfg),
		config:         cfg,
		username:       username,
		sessionNum:     sessionNum,
		command:        command,
		statusCallback: func(sessionNumber int, status string, duration time.Duration) {},
	}
}

func (r *Runner) Run(callback func(sessionNumber int, status string, duration time.Duration)) *models.StressTestResult {
	r.statusCallback = callback
	r.wg.Add(1)
	defer r.wg.Done()
	startTime := time.Now()
	result := &models.StressTestResult{
		Username:   r.username,
		TotalKasms: r.sessionNum.Value,
	}

	user, err := r.client.GetUserInfo(r.username)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get user info: %v", err))
		return result
	}
	r.UserID = user.UserID

	for i := 0; i < r.sessionNum.Value; i++ {
		kasmResult := r.createAndTestKasm(i, user.UserID)
		result.KasmResults = append(result.KasmResults, kasmResult)

		if kasmResult.ExecutionError == "" {
			result.SuccessfulKasms++
			if kasmResult.KasmID != "" {
				r.kasmsToDestroy = append(r.kasmsToDestroy, kasmResult.KasmID)
			}
		} else {
			result.FailedKasms++
			result.Errors = append(result.Errors, fmt.Sprintf("Kasm %d: %s", i, kasmResult.ExecutionError))
		}
		result.AverageStartTime += kasmResult.StartTime
	}

	result.TotalDuration = time.Since(startTime)
	if result.TotalKasms > 0 {
		result.AverageStartTime /= time.Duration(result.TotalKasms)
	}

	return result
}

func (r *Runner) createAndTestKasm(numKasms int, userID string) models.KasmResult {
	result := models.KasmResult{
		KasmNumber: numKasms + 1,
	}

	// utils.Console("Starting session %d for user %s\n", numKasms+1, r.username)
	utils.Info("Starting test for Kasm %d", numKasms+1)
	startTime := time.Now()

	// Step 1: Request Kasm
	utils.Info("Step 1: Requesting Kasm for user %s", r.username)
	kasm, err := r.client.RequestKasm(userID, r.config.DefaultImageID)
	r.statusCallback(numKasms, "Requesting Kasm", time.Since(startTime))
	if err != nil {
		utils.Error("Failed to request Kasm for user %s: %v", r.username, err)
		result.ExecutionError = fmt.Sprintf("Failed to request Kasm: %v", err)
		return result
	}

	if kasm == nil || kasm.KasmID == "" {
		result.ExecutionError = "Received empty Kasm ID from API"
		return result
	}

	result.KasmID = kasm.KasmID

	// Step 2: Wait for Kasm to be ready
	utils.Info("Step 2: Waiting for Kasm %s to be ready", kasm.KasmID)
	err = r.client.WaitForKasmReady(kasm.KasmID, userID, 10*time.Minute)
	r.statusCallback(numKasms, "Waiting for Kasm", time.Since(startTime))
	if err != nil {
		if strings.Contains(err.Error(), "stuck in 'requested' state for too long") {
			utils.Error("Kasm %s stuck in 'requested' state. Attempting to destroy and recreate.", kasm.KasmID)
			r.client.DestroyKasm(kasm.KasmID, userID)
			utils.Console("Giving the new agent a chance to catch up. Sleeiping for 5 minutes")
			time.Sleep(5 * time.Minute)
			return r.createAndTestKasm(numKasms, userID) // Recursive call to retry
		}
		utils.Error("Failed waiting for Kasm %s to be ready: %v", kasm.KasmID, err)
		result.ExecutionError = fmt.Sprintf("Failed waiting for Kasm to be ready: %v", err)
		return result
	}

	result.StartTime = time.Since(startTime)

	// Step 3: Execute command
	utils.Info("Step 3: Executing command on Kasm %s", kasm.KasmID)

	if r.command == "all" {
		// Execute CPU test
		err = r.client.ExecCommand(kasm.KasmID, userID, r.getCPUCommand())
		r.statusCallback(numKasms, "Executing command", time.Since(startTime))
		if err != nil {
			utils.Error("Failed to execute CPU command on Kasm %s: %v", kasm.KasmID, err)
			result.ExecutionError = fmt.Sprintf("Failed to execute CPU command: %v", err)
		} else {
			utils.Info("CPU command executed on Kasm %s", kasm.KasmID)
		}

		// Execute Network test
		err = r.client.ExecCommand(kasm.KasmID, userID, r.getNetworkCommand())
		r.statusCallback(numKasms, "Executing command", time.Since(startTime))
		if err != nil {
			utils.Error("Failed to execute Network command on Kasm %s: %v", kasm.KasmID, err)
			result.ExecutionError += fmt.Sprintf(" Failed to execute Network command: %v", err)
		} else {
			utils.Info("Network command executed on Kasm %s", kasm.KasmID)
		}
	} else {
		// Execute single command for other cases
		command := r.getCommandToExecute()
		err = r.client.ExecCommand(kasm.KasmID, userID, command)
		r.statusCallback(numKasms, "Executing command", time.Since(startTime))
		if err != nil {
			utils.Error("Failed to execute command on Kasm %s: %v", kasm.KasmID, err)
			result.ExecutionError = fmt.Sprintf("Failed to execute command: %v", err)
		} else {
			utils.Info("Command executed on Kasm %s", kasm.KasmID)
		}
	}

	utils.Info("Completed test for Kasm %d", numKasms+1)
	r.statusCallback(numKasms, "Completed", time.Since(startTime))
	return result
}

func (r *Runner) getCPUCommand() string {
	return "dd if=/dev/zero of=/dev/null bs=1M count=1000"
}

func (r *Runner) getNetworkCommand() string {
	return "wget -O /dev/null https://releases.ubuntu.com/22.04/ubuntu-22.04.3-live-server-amd64.iso"
}

func (r *Runner) getCommandToExecute() string {
	switch r.command {
	case "cpu":
		return r.getCPUCommand()
	case "network":
		return r.getNetworkCommand()
	default:
		return "echo 'Hello, Kasm!'"
	}
}

func (r *Runner) DestroyAllSessions() error {
	var errors []string
	for _, kasmID := range r.kasmsToDestroy {
		if err := r.client.DestroyKasm(kasmID, r.UserID); err != nil {
			utils.Info("Kasm ID: %s, User ID: %v", kasmID, r.UserID)
			utils.Error("Failed to destroy Kasm %s: %v", kasmID, err)
			errors = append(errors, fmt.Sprintf("Failed to destroy Kasm %s: %v", kasmID, err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (r *Runner) GetAutoscalingStatus() (*models.AutoscalingStatus, error) {
	// Implement this method if your API provides autoscaling information
	// For now, we'll return a placeholder
	return &models.AutoscalingStatus{
		CurrentNodes: 1,
		DesiredNodes: 1,
		PendingNodes: 0,
		MaxNodes:     10,
		CurrentLoad:  0.5,
	}, nil
}

func (r *Runner) Wait() {
	r.wg.Wait()
}
