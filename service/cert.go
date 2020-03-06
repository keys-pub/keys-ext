package service

import (
	"io/ioutil"
	"os"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/pkg/errors"
)

func certificateKey(cfg *Config, st keyring.Store, generate bool) (*keys.CertificateKey, error) {
	private, err := st.Get(cfg.AppName(), ".cert-private")
	if err != nil {
		return nil, err
	}
	public, err := st.Get(cfg.AppName(), ".cert-public")
	if err != nil {
		return nil, err
	}
	if private != nil && public != nil {
		logger.Infof("Found certificate in keyring")

		// Save public cert to filesystem too, if it doesn't exist.
		certPath, err := cfg.certPath(false)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(certPath); err == nil {
			logger.Infof("Certificate exists at %s", certPath)
		} else if os.IsNotExist(err) {
			if err := saveCertificate(cfg, string(public)); err != nil {
				return nil, errors.Wrapf(err, "failed to save cert public key")
			}
		} else {
			return nil, err
		}
		return keys.NewCertificateKey(string(private), string(public))
	}
	return generateCertificate(cfg, st)
}

// generateCertificate generates a certificate key.
func generateCertificate(cfg *Config, st keyring.Store) (*keys.CertificateKey, error) {
	logger.Infof("Generating certificate...")
	certKey, err := keys.GenerateCertificateKey("localhost", true, nil)
	if err != nil {
		return nil, err
	}

	logger.Infof("Saving certificate to keyring...")
	if err := st.Set(cfg.AppName(), ".cert-private", []byte(certKey.Private()), ""); err != nil {
		return nil, err
	}
	if err := st.Set(cfg.AppName(), ".cert-public", []byte(certKey.Public()), ""); err != nil {
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
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(certPath)
}
