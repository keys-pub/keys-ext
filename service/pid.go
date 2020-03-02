package service

import (
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
)

// checkForPID checks path for pid.
func checkForPID(path string) (int, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return -1, nil
	}
	data, err := ioutil.ReadFile(path) // #nosec
	if err != nil {
		return -1, err
	}
	n, err := strconv.Atoi(string(data))
	if err != nil {
		return -1, err
	}
	if n < 0 {
		return -1, errors.Errorf("negative pid")
	}
	return n, nil
}

var checkNoop = func() error { return nil }

// waitForPID waits for PID at path and service status.
func waitForPID(pidPath string, checkFn func() error, delay time.Duration, wait time.Duration) (int, error) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	type result struct {
		pid int
		err error
	}

	resultChan := make(chan result, 1)
	go func() {
		for {
			<-ticker.C
			pid, perr := checkForPID(pidPath)
			if perr != nil {
				resultChan <- result{pid: -1, err: perr}
				return
			}
			if checkFn != nil {
				if err := checkFn(); err != nil {
					resultChan <- result{pid: -1, err: err}
					return
				}
			}
			// If PID not found, continue
			if pid == -1 {
				continue
			}
			resultChan <- result{pid: pid}
			return
		}
	}()

	select {
	case res := <-resultChan:
		return res.pid, res.err
	case <-time.After(wait):
		return -1, errors.Errorf("timed out waiting for pid")
	}
}

// waitForProcessExit waits for PID to exit.
func waitForProcessExit(pid int, delay time.Duration, wait time.Duration) error {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	type result struct {
		err error
	}

	resultChan := make(chan result, 1)
	go func() {
		for {
			<-ticker.C
			process, findErr := ps.FindProcess(pid)
			if findErr != nil {
				resultChan <- result{err: findErr}
				return
			}
			if process == nil {
				resultChan <- result{}
				return
			}
		}
	}()

	select {
	case res := <-resultChan:
		return res.err
	case <-time.After(wait):
		return errors.Errorf("timed out waiting for pid to exit")
	}
}
