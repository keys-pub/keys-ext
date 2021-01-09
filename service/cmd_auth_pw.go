package service

import (
	"context"
	"fmt"
	"os"
)

func passwordAuthSetup(ctx context.Context, client *Client, clientName string, password string) (string, error) {
	if len(password) == 0 {
		fmt.Fprintf(os.Stderr, "OK, let's create a password.\n")
		p, err := readVerifyPassword("Create a password:")
		if err != nil {
			return "", err
		}
		password = p
	}

	if _, err := client.RPCClient().AuthSetup(ctx, &AuthSetupRequest{
		Secret: password,
		Type:   PasswordAuth,
	}); err != nil {
		return "", err
	}

	unlockResp, err := client.RPCClient().AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}

	return unlockResp.AuthToken, nil
}

func passwordAuthUnlock(ctx context.Context, client *Client, clientName string, password string) (string, error) {
	if len(password) == 0 {
		p, err := readPassword("Enter your password:", false)
		if err != nil {
			return "", err
		}
		password = p
	}

	unlock, err := client.RPCClient().AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Secret: password,
		Type:   PasswordAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}
	return unlock.AuthToken, nil
}

func passwordAuthProvision(ctx context.Context, client *Client, password string) error {
	if len(password) == 0 {
		fmt.Fprintf(os.Stderr, "OK, let's create a password.\n")
		p, err := readVerifyPassword("Create a password:")
		if err != nil {
			return err
		}
		password = p
	}

	if _, err := client.RPCClient().AuthProvision(ctx, &AuthProvisionRequest{
		Secret: password,
		Type:   PasswordAuth,
	}); err != nil {
		return err
	}

	return nil
}
