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
	"syscall"
	"time"
)

func main() {
	pattern_flag := flag.String("p", "", "Pattern to match (required)")
	timeout_flag := flag.Duration("t", 3*time.Second, "Timeout duration. Set to 0 for no timeout.")
	kill_flag := flag.Bool("k", true, "Whether the process should be terminated on pattern matched.")

	flag.Usage = func() {
		fmt.Print("usage:", os.Args[0], "[flags] <command>\n\n")
		fmt.Println("flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *pattern_flag == "" {
		warn("Error: Pattern flag is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if flag.NArg() < 1 {
		warn("Error: Command argument is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	command := strings.Join(flag.Args(), " ")

	err := executeCommandUntilMatch(command, *pattern_flag, *timeout_flag, *kill_flag)
	if err != nil {
		warn("Error:", err)
		os.Exit(1)
	}
}

func executeCommandUntilMatch(command string, pattern string, timeout time.Duration, do_kill bool) error {
	cmd := exec.Command(command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout_done := make(chan bool)
	stderr_done := make(chan bool)
	pattern_match := make(chan string)

	go listenForPatternMatches(stdout, "stdout", pattern, pattern_match, stdout_done)
	go listenForPatternMatches(stderr, "stderr", pattern, pattern_match, stderr_done)

	if err := cmd.Start(); err != nil {
		return err
	}

	if timeout == 0 {
		select {
		case line := <-pattern_match:
			return returnMatch(line, cmd, do_kill)
		case <-stdout_done:
			return returnNoMatch()
		case <-stderr_done:
			return returnNoMatch()
		}
	} else {
		select {
		case line := <-pattern_match:
			return returnMatch(line, cmd, do_kill)
		case <-stdout_done:
			return returnNoMatch()
		case <-stderr_done:
			return returnNoMatch()
		case <-time.After(timeout):
			return returnTimeout(cmd, do_kill)
		}
	}
}

func listenForPatternMatches(reader io.ReadCloser, scanner_type string, pattern string, pattern_match chan string, done chan bool) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		log(scanner_type, ":", line)

		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			warn("Regex matching error", scanner_type, ":", err)
		}
		if matched {
			pattern_match <- line
		}
	}

	if err := scanner.Err(); err != nil {
		warn("Scanner error", scanner_type, ":", err)
	}

	done <- true
}

const (
	ERROR_PATTERN_NOT_FOUND = "command completed without a match"
	ERROR_TIMEOUT           = "timeout reached, pattern not matched"
)

func maybeKillProcess(do_kill bool, cmd *exec.Cmd) {
	if do_kill {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
			warn("failed attempting to kill running command", err)
		}
		log("killed the running command")
	}
}

func returnTimeout(cmd *exec.Cmd, do_kill bool) error {
	maybeKillProcess(do_kill, cmd)
	return errors.New(ERROR_TIMEOUT)
}

func returnMatch(line string, cmd *exec.Cmd, do_kill bool) error {
	maybeKillProcess(do_kill, cmd)
	log("pattern matched", line)
	return nil
}

func returnNoMatch() error {
	return errors.New(ERROR_PATTERN_NOT_FOUND)
}

func warn(things ...any) {
	fmt.Fprintln(os.Stderr, things...)
}

func log(things ...any) {
	fmt.Println(things...)
}
