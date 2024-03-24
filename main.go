package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
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
		fmt.Print("usage:", os.Args[0], "[flags] <command>\n\n")
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
		fmt.Fprintln(os.Stderr, "Error:", err)
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
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	stdout_done := make(chan bool)
	stderr_done := make(chan bool)
	pattern_match := make(chan string)

	go listenForPatternMatches(stdout, "stdout", pattern, pattern_match, stdout_done)
	go listenForPatternMatches(stderr, "stderr", pattern, pattern_match, stderr_done)

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if timeout == 0 {
		select {
		case line := <-pattern_match:
			return returnMatch(line, cmd)
		case <-stdout_done:
			return returnNoMatch(stderr_done)
		case <-stderr_done:
			return returnNoMatch(stdout_done)
		}
	} else {
		select {
		case line := <-pattern_match:
			return returnMatch(line, cmd)
		case <-stdout_done:
			return returnNoMatch(stderr_done)
		case <-stderr_done:
			return returnNoMatch(stdout_done)
		case <-time.After(timeout):
			return returnTimeout(cmd)
		}
	}
}

func listenForPatternMatches(reader io.ReadCloser, scannerType string, pattern string, pattern_match chan string, done chan bool) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(scannerType, ":", line)
		if matched, _ := MatchPattern(line, pattern); matched {
			pattern_match <- line
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Scanner error", scannerType, ":", err)
	}

	done <- true
}

func MatchPattern(text string, pattern string) (bool, error) {
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func returnTimeout(cmd *exec.Cmd) (string, error) {
	cmd.Process.Kill()
	return "", errors.New("timeout reached, pattern not matched")
}

func returnMatch(line string, cmd *exec.Cmd) (string, error) {
	cmd.Process.Kill()
	return line, nil
}

func returnNoMatch(exitSignalFromOtherListener chan bool) (string, error) {
	<-exitSignalFromOtherListener
	return "", errors.New("command completed without a match")
}
