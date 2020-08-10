package service

import (
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// GenerateCertificate generates a certificate key and saves it to the support dir.
func GenerateCertificate(cfg *Config, save bool) (*keys.CertificateKey, error) {
	logger.Infof("Generating certificate...")
	certKey, err := keys.GenerateCertificateKey("localhost", true, nil)
	if err != nil {
		return nil, err
	}
	if save {
		if err := saveCertificate(cfg, certKey.Public()); err != nil {
			return nil, errors.Wrapf(err, "failed to save cert public key")
		}
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
	return ioutil.WriteFile(certPath, []byte(cert), filePerms)
}

// loadCertificate returns public certificate PEM from the filesystem.
func loadCertificate(cfg *Config) (string, error) {
	certPath, err := cfg.certPath(false)
	if err != nil {
		return "", err
	}
	exists, err := pathExists(certPath)
	if err != nil {
		return "", err
	}
	if !exists {
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

// DeleteCertificate removes saved certificate.
func DeleteCertificate(cfg *Config) error {
	certPath, err := cfg.certPath(false)
	if err != nil {
		return err
	}
	exists, err := pathExists(certPath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.Remove(certPath)
}
