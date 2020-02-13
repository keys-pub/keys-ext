package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

func readPassword(prompt string) (string, error) {
	if !terminal.IsTerminal(int(syscall.Stdin)) {
		return "", errors.Errorf("failed to read password from terminal: not a terminal or terminal not supported, use -password option")
	}

	// Get the initial state of the terminal.
	initialTermState, err := terminal.GetState(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	// Restore it in the event of an interrupt.
	// CITATION: Konstantin Shaposhnikov - https://groups.google.com/forum/#!topic/golang-nuts/kTVAbtee9UA
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(int(syscall.Stdin), initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Fprintf(os.Stderr, prompt)
	p, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintf(os.Stderr, "\n")

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	if err != nil {
		return "", err
	}

	// Return the password as a string.
	return string(p), nil
}

func readVerifyPassword(prompt string) (string, error) {
	password, err := readPassword(prompt)
	if err != nil {
		return "", err
	}
	password2, err := readPassword("Re-enter the password:")
	if err != nil {
		return "", err
	}

	if password != password2 {
		return "", errors.Errorf("passwords don't match")
	}

	return password, nil
}
