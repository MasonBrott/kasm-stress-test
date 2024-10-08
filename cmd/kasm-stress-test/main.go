package main

import (
	"bufio"
	"flag"
	"fmt"
	"kasm-stress-test/internal/config"
	"kasm-stress-test/internal/models"
	"kasm-stress-test/internal/stress"
	"kasm-stress-test/internal/utils"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type SessionStatus struct {
	Status   string
	Duration time.Duration
}

var (
	sessionStatuses map[string][]SessionStatus
	statusMutex     sync.Mutex
	allRunners      []*stress.Runner
	allResults      []*models.StressTestResult
	resultsMutex    sync.Mutex
	startTime       time.Time
	lastOutput      []string
	updateChan      chan struct{}
)

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func updateSessionStatus(username string, sessionNumber int, status string, duration time.Duration) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	if sessionStatuses[username] == nil {
		sessionStatuses[username] = make([]SessionStatus, 0)
	}
	for len(sessionStatuses[username]) <= sessionNumber {
		sessionStatuses[username] = append(sessionStatuses[username], SessionStatus{})
	}
	sessionStatuses[username][sessionNumber] = SessionStatus{Status: status, Duration: duration}
}

func clearScreen() {
	fmt.Print("\033[2J")
	moveCursorToTop()
}

func moveCursorToTop() {
	fmt.Print("\033[H")
}

func updateLine(content string) {
	fmt.Printf("\033[K%s", content) // Clear line and print new content
}

func updateDisplay() {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	var newOutput []string

	elapsedTime := time.Since(startTime)
	newOutput = append(newOutput, fmt.Sprintf("Kasm Stress Test Status - Elapsed Time: %s", formatDuration(elapsedTime)))
	newOutput = append(newOutput, "")

	// Create a sorted list of usernames
	var usernames []string
	for username := range sessionStatuses {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	// Iterate through sorted usernames
	for _, username := range usernames {
		sessions := sessionStatuses[username]
		newOutput = append(newOutput, fmt.Sprintf("Starting %d sessions for user %s", len(sessions), username))
		for i, session := range sessions {
			newOutput = append(newOutput, fmt.Sprintf("Session %d:", i+1))
			newOutput = append(newOutput, fmt.Sprintf("    Status - %s", session.Status))
			newOutput = append(newOutput, fmt.Sprintf("    Duration - %s", formatDuration(session.Duration)))
		}
		newOutput = append(newOutput, "")
	}

	moveCursorToTop()
	for i, line := range newOutput {
		if i >= len(lastOutput) || line != lastOutput[i] {
			updateLine(line)
		}
		fmt.Print("\n")
	}

	// Clear any remaining lines from the previous output
	for i := len(newOutput); i < len(lastOutput); i++ {
		updateLine("")
		fmt.Print("\n")
	}

	lastOutput = newOutput
}

func main() {
	err := utils.InitLoggers()
	if err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}
	defer utils.CloseLogFile()

	var usernames utils.StringSliceFlag
	flag.Var(&usernames, "u", "Username to use (can be specified multiple times)")
	flag.Var(&usernames, "username", "Username to use (can be specified multiple times)")

	var sessionNum utils.IntFlag
	flag.Var(&sessionNum, "n", "Number of Kasm Sessions to start for each username specified")
	flag.Var(&sessionNum, "number", "Number of Kasm Sessions to start for each username specified")

	var command string
	flag.StringVar(&command, "c", "all", "Command to run: 'cpu', 'network', or 'all' (default)")
	flag.StringVar(&command, "command", "all", "Command to run: 'cpu', 'network', or 'all' (default)")

	flag.Parse()

	if len(usernames) == 0 {
		log.Fatal("At least one username is required")
	}

	if len(sessionNum.String()) == 0 {
		log.Fatal("Please provide the number of sessions to start")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sessionStatuses = make(map[string][]SessionStatus)
	updateChan = make(chan struct{}, 100)

	startTime = time.Now()

	// Clear the screen and hide the cursor
	fmt.Print("\033[2J\033[?25l")
	defer fmt.Print("\033[?25h") // Show the cursor when done

	// Start a goroutine to update the display
	stopChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Update elapsed time
				updateDisplay()
			case <-updateChan:
				// Update session statuses
				updateDisplay()
			case <-stopChan:
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for _, username := range usernames {
		wg.Add(1)
		go func(username string) {
			defer wg.Done()
			runner := stress.NewRunner(cfg, username, sessionNum, command)
			allRunners = append(allRunners, runner)
			for i := 0; i < sessionNum.Value; i++ {
				updateSessionStatus(username, i, "Starting", 0)
				updateChan <- struct{}{}
			}
			results := runner.Run(func(sessionNumber int, status string, duration time.Duration) {
				updateSessionStatus(username, sessionNumber, status, duration)
				updateChan <- struct{}{}
				time.Sleep(100 * time.Millisecond) // Short delay after each status update
			})
			resultsMutex.Lock()
			allResults = append(allResults, results)
			resultsMutex.Unlock()
		}(username)
	}

	wg.Wait()
	close(stopChan)

	// Clear the screen one last time before showing results
	clearScreen()

	// Process and print all results
	utils.Console("\n--- Stress Test Results ---\n")
	for _, result := range allResults {
		utils.Console("\nResults for user: %s\n", result.Username)
		utils.Console("Total Kasms created: %d\n", result.TotalKasms)
		utils.Console("Successful Kasms: %d\n", result.SuccessfulKasms)
		utils.Console("Failed Kasms: %d\n", result.FailedKasms)
		utils.Console("Average start time: %.2f seconds\n", result.AverageStartTime.Seconds())
		utils.Console("Total duration: %.2f seconds\n", result.TotalDuration.Seconds())

		if len(result.Errors) > 0 {
			utils.Console("Errors encountered:\n")
			for _, err := range result.Errors {
				utils.Console("  - %s\n", err)
			}
		}

		utils.Console("\nDetailed Kasm Results:\n")
		for _, kasmResult := range result.KasmResults {
			utils.Console("  Kasm #%d:\n", kasmResult.KasmNumber)
			utils.Console("    Start time: %.2f seconds\n", kasmResult.StartTime.Seconds())
			if kasmResult.ExecutionError != "" {
				utils.Console("    Error: %s\n", kasmResult.ExecutionError)
			} else {
				utils.Console("    Status: Success\n")
			}
		}
		utils.Console("%s\n", strings.Repeat("-", 30))
	}
	utils.Info("Stress test completed")

	// Prompt user to press Enter before destroying sessions
	utils.Console("\nPress Enter to destroy sessions and complete the test\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	utils.Console("Destroying Sessions...\n")

	// Destroy all Sessions
	var destroyErrors []error
	for _, runner := range allRunners {
		if err := runner.DestroyAllSessions(); err != nil {
			utils.Error("Error destroying Kasms: %v", err)
			destroyErrors = append(destroyErrors, err)
		}
	}

	if len(destroyErrors) == 0 {
		fmt.Println("\nAll Kasm sessions have been successfully destroyed. Test complete.")
	} else {
		utils.Error("\nTest complete, but some Kasm sessions could not be destroyed.")
	}
}
