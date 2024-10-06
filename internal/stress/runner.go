package stress

import (
	"fmt"
	"strings"
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
}

func NewRunner(cfg *config.Config, username string, sessionNum utils.IntFlag, command string) *Runner {
	return &Runner{
		client:     api.NewClient(cfg),
		config:     cfg,
		username:   username,
		sessionNum: sessionNum,
		command:    command,
	}
}

func (r *Runner) Run() *models.StressTestResult {
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
			r.kasmsToDestroy = append(r.kasmsToDestroy, kasmResult.KasmID)
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
		KasmNumber: numKasms,
	}

	utils.Info("Starting test for Kasm %d", numKasms)
	startTime := time.Now()

	// Step 1: Request Kasm
	utils.Info("Step 1: Requesting Kasm for user %s", r.username)
	kasm, err := r.client.RequestKasm(userID, r.config.DefaultImageID)
	if err != nil {
		utils.Error("Failed to request Kasm for user %s: %v", r.username, err)
		return result
	}

	result.KasmID = kasm.KasmID

	// Step 2: Wait for Kasm to be ready
	utils.Info("Step 2: Waiting for Kasm %s to be ready", kasm.KasmID)
	err = r.client.WaitForKasmReady(kasm.KasmID, userID, 5*time.Minute)
	if err != nil {
		utils.Error("Failed waiting for Kasm %s to be ready: %v", kasm.KasmID, err)
		return result
	}

	result.StartTime = time.Since(startTime)

	// Step 3: Execute command
	utils.Info("Step 3: Executing command on Kasm %s", kasm.KasmID)
	command := r.getCommandToExecute()
	err = r.client.ExecCommand(kasm.KasmID, userID, command)
	if err != nil {
		utils.Error("Failed to execute command on Kasm %s: %v", kasm.KasmID, err)
		result.ExecutionError = fmt.Sprintf("Failed to execute command: %v", err)
	} else {
		utils.Info("Command executed on Kasm %s", kasm.KasmID)
	}

	utils.Info("Completed test for Kasm %d", numKasms)
	return result
}

func (r *Runner) getCommandToExecute() string {
	switch r.command {
	case "cpu":
		// 1000 MB of writes to /dev/null
		return "dd if=/dev/zero of=/dev/null bs=1M count=1000"
	case "network":
		// Downloads a 10MB file 10 times without saving file
		return "for i in {1..10}; do wget -O /dev/null http://speedtest.wdc01.softlayer.com/downloads/test10.zip; done"
	case "all":
		return "(dd if=/dev/zero of=/dev/null bs=1M count=1000 &) && (for i in {1..10}; do wget -O /dev/null http://speedtest.wdc01.softlayer.com/downloads/test10.zip; done &) && wait"
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
