package vault

import (
	"strconv"
	"time"

	"github.com/keys-pub/keys/tsutil"

	"github.com/keys-pub/keys/ds"
)

func (v *Vault) index() (int64, error) {
	return v.getConfigInt64("index")
}

func (v *Vault) setIndex(n int64) error {
	return v.setConfigInt64("index", n)
}

func (v *Vault) autoSyncDisabled() (bool, error) {
	return v.getConfigBool("autoSyncDisabled")
}

func (v *Vault) setAutoSyncDisabled(b bool) error {
	return v.setConfigBool("autoSyncDisabled", b)
}

func (v *Vault) lastSync() (time.Time, error) {
	return v.getConfigTime("lastSync")
}

func (v *Vault) setLastSync(tm time.Time) error {
	return v.setConfigTime("lastSync", tm)
}

func (v *Vault) setConfig(key string, value string) error {
	if err := v.store.Set(ds.Path("db", key), []byte(value)); err != nil {
		return err
	}
	return nil
}

func (v *Vault) setConfigInt64(key string, n int64) error {
	return v.setConfig(key, strconv.FormatInt(n, 10))
}

func (v *Vault) setConfigBool(key string, b bool) error {
	s := "0"
	if b {
		s = "1"
	}
	return v.setConfig(key, s)
}

func (v *Vault) setConfigTime(key string, t time.Time) error {
	return v.setConfigInt64(key, tsutil.Millis(t))
}

func (v *Vault) getConfig(key string) (string, error) {
	b, err := v.store.Get(ds.Path("db", key))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (v *Vault) getConfigInt64(key string) (int64, error) {
	s, err := v.getConfig(key)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (v *Vault) getConfigTime(key string) (time.Time, error) {
	n, err := v.getConfigInt64(key)
	if err != nil {
		return time.Time{}, err
	}
	if n == 0 {
		return time.Time{}, nil
	}
	return tsutil.ParseMillis(n), nil
}

func (v *Vault) getConfigBool(key string) (bool, error) {
	s, err := v.getConfig(key)
	if err != nil {
		return false, err
	}
	if s == "1" {
		return true, nil
	}
	return false, nil
}
