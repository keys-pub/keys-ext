package service

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
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

func autostart(env *Env) error {
	logger.Infof("Autostart")
	if err := startProcess(env); err != nil {
		if err == errAlreadyRunning {
			logger.Debugf("Already running")
			return nil
		}
		return err
	}
	logger.Debugf("Autostarted")
	return waitForStart(env)
}

func clearPID(env *Env) error {
	pidPath, err := env.AppPath("pid", false)
	if err != nil {
		return err
	}
	if err := removeFile(pidPath); err != nil {
		return err
	}
	return nil
}

func waitForStart(env *Env) error {
	pidPath, err := env.AppPath("pid", false)
	if err != nil {
		return err
	}
	logger.Debugf("Waiting for pid: %s", pidPath)
	_, perr := waitForPID(pidPath, nil, time.Second, 20*time.Second)
	if perr != nil {
		return perr
	}
	return nil
}

var errNotRunning = errors.New("not running")
var errAlreadyRunning = errors.New("already running")

func startProcess(env *Env) error {
	logger.Debugf("Start process")
	ps, err := findProcessByName("keysd")
	if err != nil {
		return err
	}
	if ps != nil {
		return errAlreadyRunning
	}

	binPath := defaultServicePath()
	appName := env.AppName()
	logPath, err := env.LogsPath("keysd.log", true)
	if err != nil {
		return err
	}

	args := []string{
		"-app", appName,
		"-log-path", logPath,
	}
	logger.Debugf("Starting %s %s", binPath, strings.Join(args, " "))
	cmd := exec.Command(binPath, args...) // #nosec
	return cmd.Start()
}

func stopProcess(env *Env) error {
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
