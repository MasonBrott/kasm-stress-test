package main

import (
	"flag"
	"fmt"
	"kasm-stress-test/internal/config"
	"kasm-stress-test/internal/models"
	"kasm-stress-test/internal/stress"
	"kasm-stress-test/internal/utils"
	"log"
	"strings"
)

func main() {
	utils.InitLoggers()
	var usernames utils.StringSliceFlag
	flag.Var(&usernames, "u", "Usernames to use (can be specified multiple times)")

	var kasmRange utils.IntRangeFlag
	flag.Var(&kasmRange, "kasm-range", "Range of Kasm instances to create (e.g., 5-10)")

	flag.Parse()

	if len(usernames) == 0 {
		log.Fatal("At least one username is required")
	}

	if kasmRange.Min == 0 || kasmRange.Max == 0 {
		log.Fatal("Invalid Kasm range")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	utils.Info("Starting stress test for %d users", len(usernames))

	var allResults []*models.StressTestResult

	for _, username := range usernames {
		runner := stress.NewRunner(cfg, username, kasmRange)
		results := runner.Run()
		allResults = append(allResults, results)
	}

	// Process and print all results
	fmt.Println("\n--- Stress Test Results ---")
	for _, result := range allResults {
		fmt.Printf("\nResults for user: %s\n", result.Username)
		fmt.Printf("Total Kasms created: %d\n", result.TotalKasms)
		fmt.Printf("Successful Kasms: %d\n", result.SuccessfulKasms)
		fmt.Printf("Failed Kasms: %d\n", result.FailedKasms)
		fmt.Printf("Average start time: %.2f seconds\n", result.AverageStartTime.Seconds())

		if len(result.Errors) > 0 {
			fmt.Println("Errors encountered:")
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}

		fmt.Println("\nDetailed Kasm Results:")
		for _, kasmResult := range result.KasmResults {
			fmt.Printf("  Kasm #%d:\n", kasmResult.KasmNumber)
			fmt.Printf("    Start time: %.2f seconds\n", kasmResult.StartTime.Seconds())
			if kasmResult.ExecutionError != "" {
				fmt.Printf("    Error: %s\n", kasmResult.ExecutionError)
			} else {
				fmt.Println("    Status: Success")
			}
		}
		fmt.Println(strings.Repeat("-", 30))
	}
	utils.Info("Stress test completed")
}
