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
	"strings"
)

func main() {
	err := utils.InitLoggers()
	if err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}
	defer utils.CloseLogFile()

	utils.InitLoggers()
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

	utils.Info("Starting stress test for %d users", len(usernames))

	var allResults []*models.StressTestResult
	var allRunners []*stress.Runner

	for _, username := range usernames {
		runner := stress.NewRunner(cfg, username, sessionNum, command)
		allRunners = append(allRunners, runner)
		results := runner.Run()
		allResults = append(allResults, results)
	}

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
