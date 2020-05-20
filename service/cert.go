package service

import (
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// generateCertificate generates a certificate key.
func generateCertificate(cfg *Config) (*keys.CertificateKey, error) {
	logger.Infof("Generating certificate...")
	certKey, err := keys.GenerateCertificateKey("localhost", true, nil)
	if err != nil {
		return nil, err
	}
	if err := saveCertificate(cfg, certKey.Public()); err != nil {
		return nil, errors.Wrapf(err, "failed to save cert public key")
	}
	return certKey, nil
}

// saveCertificate saves public certificate PEM data to the filesystem.
func saveCertificate(cfg *Config, cert string) error {
	certPath, err := cfg.certPath(true)
	if err != nil {
		return err
	}
	logger.Infof("Saving certificate PEM %s", certPath)
	return ioutil.WriteFile(certPath, []byte(cert), 0600)
}

// loadCertificate returns public certificate PEM from the filesystem.
func loadCertificate(cfg *Config) (string, error) {
	certPath, err := cfg.certPath(false)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", nil
	}
	logger.Debugf("Loading certificate %s", certPath)
	b, err := ioutil.ReadFile(certPath) // #nosec
	if err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", errors.Errorf("certificate is not valid utf8")
	}
	return string(b), nil
}

// deleteCertificate removes saved certificate.
func deleteCertificate(cfg *Config) error {
	certPath, err := cfg.certPath(false)
	if err != nil {
		return err
	}
	if _, err := os.Stat(certPath); err == nil {
		return os.Remove(certPath)
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}
