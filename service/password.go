package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/term"
)

func readPassword(prompt string, allowEmpty bool) (string, error) {
	if !term.IsTerminal(int(syscall.Stdin)) {
		return "", errors.Errorf("failed to read password from terminal: not a terminal or terminal not supported, use -password option")
	}

	// Bug with windows
	// https://github.com/golang/go/issues/36609
	// if runtime.GOOS == "windows" {
	// 	return "", errors.Errorf("temporarily unsupported on windows, use -password option")
	// }

	// Get the initial state of the terminal.
	initialTermState, err := term.GetState(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	// Restore it in the event of an interrupt.
	// CITATION: Konstantin Shaposhnikov - https://groups.google.com/forum/#!topic/golang-nuts/kTVAbtee9UA
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill) //nolint
	go func() {
		<-c
		_ = term.Restore(int(syscall.Stdin), initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Fprint(os.Stderr, prompt)
	p, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintf(os.Stderr, "\n")

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	if err != nil {
		return "", err
	}

	if len(p) == 0 && !allowEmpty {
		return "", errors.Errorf("empty password")
	}

	// Return the password as a string.
	return string(p), nil
}

func readVerifyPassword(prompt string) (string, error) {
	password, err := readPassword(prompt, false)
	if err != nil {
		return "", err
	}
	password2, err := readPassword("Re-enter the password:", false)
	if err != nil {
		return "", err
	}

	if password != password2 {
		return "", errors.Errorf("passwords don't match")
	}

	return password, nil
}
