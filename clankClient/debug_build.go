package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Change to clankClient directory
	clankClientDir := "C:\\Users\\Celot\\Documents\\Projects\\corruptionTracker\\clankClient"
	err := os.Chdir(clankClientDir)
	if err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}
	
	fmt.Printf("Working directory: %s\n", clankClientDir)
	
	// Run go mod tidy
	fmt.Println("Running 'go mod tidy'...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("go mod tidy failed: %v\n", err)
		fmt.Printf("Output: %s\n", tidyOutput)
	} else {
		fmt.Println("go mod tidy completed successfully")
	}
	
	// Try to build
	fmt.Println("\nAttempting to build...")
	buildCmd := exec.Command("go", "build", ".\\cmd\\clankClient\\main.go")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Build failed: %v\n", err)
		fmt.Printf("Build errors:\n%s\n", buildOutput)
		
		// If build failed, let's analyze the errors and continue
		analyzeAndFix(string(buildOutput))
	} else {
		fmt.Println("Build successful!")
		
		// Try to run
		fmt.Println("Running the application...")
		runCmd := exec.Command("go", "run", ".\\cmd\\clankClient\\main.go")
		runOutput, err := runCmd.CombinedOutput()
		fmt.Printf("Run output: %s\n", runOutput)
		if err != nil {
			fmt.Printf("Run failed: %v\n", err)
		}
	}
}

func analyzeAndFix(buildOutput string) {
	fmt.Println("Analyzing build errors...")
	
	lines := strings.Split(buildOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "unknown field") || strings.Contains(line, "undefined:") {
			fmt.Printf("Error to fix: %s\n", line)
		}
	}
}
