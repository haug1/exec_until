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
	timeoutFlag := flag.Duration("t", 3*time.Second, "Timeout duration. Set to 0 for no timeout.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <command>\n\n", os.Args[0])
		fmt.Println("flags:")
		flag.PrintDefaults()
	}

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

	output, err := ExecuteCommandUntilMatch(command, *patternFlag, *timeoutFlag)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("pattern matched:", output)
}

func ExecuteCommandUntilMatch(command string, pattern string, timeout time.Duration) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(stdout)

	if err := cmd.Start(); err != nil {
		return "", err
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

	if timeout == 0 {
		select {
		case line := <-pattern_match:
			return RETURN_MATCH(line, cmd)
		case <-done:
			return RETURN_NO_MATCH()
		}
	} else {
		select {
		case line := <-pattern_match:
			return RETURN_MATCH(line, cmd)
		case <-done:
			return RETURN_NO_MATCH()
		case <-time.After(timeout):
			return RETURN_TIMEOUT(cmd)
		}
	}
}

func RETURN_TIMEOUT(cmd *exec.Cmd) (string, error) {
	cmd.Process.Kill()
	return "", errors.New("timeout reached, pattern not matched")
}

func RETURN_MATCH(line string, cmd *exec.Cmd) (string, error) {
	cmd.Process.Kill()
	return line, nil
}

func RETURN_NO_MATCH() (string, error) {
	return "", errors.New("command completed without a match")
}

func MatchPattern(text string, pattern string) (bool, error) {
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false, err
	}
	return matched, nil
}
