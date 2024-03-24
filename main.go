package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	patternFlag := flag.String("p", "", "Pattern to match (required)")
	timeoutFlag := flag.Duration("t", 3*time.Second, "Timeout duration in seconds")

	flag.Parse()

	if *patternFlag == "" {
		fmt.Println("Error: Pattern flag is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: Command argument is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	command := strings.Join(flag.Args(), " ")

	output, command_completed, err := ExecuteCommandUntilMatch(command, *patternFlag, *timeoutFlag)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if output != "" {
		fmt.Println("pattern matched:", output)
	} else if command_completed {
		fmt.Println("command completed without a match")
	}
}

func ExecuteCommandUntilMatch(command string, pattern string, timeout time.Duration) (string, bool, error) {
	cmd := exec.Command("bash", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", false, err
	}

	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return "", false, err
	}

	done := make(chan bool)
	pattern_match := make(chan string)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if matched, _ := MatchPattern(line, pattern); matched {
				pattern_match <- line
				return
			}
		}
		done <- true
	}()

	select {
	case line := <-pattern_match:
		cmd.Process.Kill()
		return line, false, nil
	case <-done:
		return "", true, nil
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", false, errors.New("timeout reached, pattern not matched")
	}
}

func MatchPattern(text string, pattern string) (bool, error) {
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false, err
	}
	return matched, nil
}