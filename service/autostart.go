package service

import (
	"os"
	"os/exec"
	"runtime"
	strings "strings"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
)

func findProcessByName(name string) (ps.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		exe := p.Executable()
		if exe == name || (runtime.GOOS == "windows" && exe == name+".exe") {
			logger.Debugf("Found process: %+v", p)
			return p, nil
		}
	}
	logger.Debugf("Process not found (%s)", name)
	return nil, nil
}

func autostart(cfg *Config) error {
	if err := startProcess(cfg); err != nil {
		if err == errAlreadyRunning {
			logger.Debugf("Already running")
			return nil
		}
		return err
	}
	logger.Debugf("Autostarted")
	return waitForStart(cfg)
}

func waitForStart(cfg *Config) error {
	pidPath, err := cfg.AppPath("pid", false)
	if err != nil {
		return err
	}
	logger.Debugf("Waiting for pid: %s", pidPath)
	_, perr := waitForPID(pidPath, nil, time.Second, 10*time.Second)
	if perr != nil {
		return perr
	}
	return nil
}

var errNotRunning = errors.New("not running")
var errAlreadyRunning = errors.New("already running")

func restartProcess(cfg *Config) error {
	logger.Debugf("Restart process")
	if err := stopProcess(cfg); err != nil {
		if err != errNotRunning {
			return err
		}
	}
	return autostart(cfg)
}

func startProcess(cfg *Config) error {
	logger.Debugf("Start process")
	ps, err := findProcessByName("keysd")
	if err != nil {
		return err
	}
	if ps != nil {
		return errAlreadyRunning
	}

	binPath := defaultServicePath()
	appName := cfg.AppName()
	logPath, err := cfg.LogsPath("keysd.log", true)
	if err != nil {
		return err
	}

	args := []string{
		"-app", appName,
		"-log-path", logPath,
	}
	logger.Debugf("Starting %s %s", binPath, strings.Join(args, " "))
	cmd := exec.Command(binPath, args...)
	return cmd.Start()
}

func stopProcess(cfg *Config) error {
	logger.Debugf("Stop process")
	ps, err := findProcessByName("keysd")
	if err != nil {
		return err
	}
	if ps == nil {
		return errNotRunning
	}
	process, findErr := os.FindProcess(ps.Pid())
	if findErr != nil {
		return findErr
	}
	if process == nil {
		logger.Debugf("Process pid not found %d", ps.Pid())
		return errNotRunning
	}
	logger.Debugf("Killing process %d", ps.Pid())
	if err := process.Kill(); err != nil {
		return err
	}
	logger.Debugf("Waiting for process %d to exit", ps.Pid())
	if err := waitForProcessExit(ps.Pid(), time.Millisecond*500, time.Second*10); err != nil {
		return err
	}
	return nil
}
