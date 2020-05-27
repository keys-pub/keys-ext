package service

import (
	"context"
	"fmt"
	"os"
)

func fido2AuthSetup(ctx context.Context, client *Client, clientName string, pin string, pinRequired bool) (string, error) {
	if pinRequired && len(pin) == 0 {
		p, err := readPassword("Enter your PIN:")
		if err != nil {
			return "", err
		}
		pin = p
	}

	fmt.Fprintln(os.Stderr, "Creating credential, you may need to authorize on your device...")
	if _, err := client.KeysClient().AuthSetup(ctx, &AuthSetupRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
	}); err != nil {
		return "", err
	}

	fmt.Fprintln(os.Stderr, "Getting credential, you may need to authorize on your device (again)...")
	unlockResp, err := client.KeysClient().AuthUnlock(ctx, &AuthUnlockRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}

	return unlockResp.AuthToken, nil
}

func fido2AuthUnlock(ctx context.Context, client *Client, clientName string, pin string, pinRequired bool) (string, error) {
	if pinRequired && len(pin) == 0 {
		p, err := readPassword("Enter your PIN:")
		if err != nil {
			return "", err
		}
		pin = p
	}

	fmt.Fprintln(os.Stderr, "Getting credential, you may need to authorize on your device...")
	unlock, err := client.KeysClient().AuthUnlock(context.TODO(), &AuthUnlockRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Client: clientName,
	})
	if err != nil {
		return "", err
	}
	return unlock.AuthToken, nil
}

func fido2AuthProvision(ctx context.Context, client *Client, clientName string, pin string, setup bool) error {
	if setup {
		fmt.Fprintln(os.Stderr, "Let's create a credential on your device, you may need to interact with the key...")
	} else {
		fmt.Fprintln(os.Stderr, "Getting the credential from your device, you may need to interact with the key (again)...")
	}
	if _, err := client.KeysClient().AuthProvision(ctx, &AuthProvisionRequest{
		Secret: pin,
		Type:   FIDO2HMACSecretAuth,
		Setup:  setup,
	}); err != nil {
		return err
	}

	return nil
}
