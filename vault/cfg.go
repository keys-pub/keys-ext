package vault

import (
	"fmt"
	"strconv"
	"time"

	"github.com/keys-pub/keys/tsutil"

	"github.com/keys-pub/keys/docs"
)

func (v *Vault) setValue(key string, value []byte) error {
	path := docs.Path(key)
	if value == nil {
		if _, err := v.store.Delete(path); err != nil {
			return err
		}
		return nil
	}
	return v.store.Set(path, value)
}

func (v *Vault) setInt64(key string, n int64) error {
	if n == 0 {
		return v.setValue(key, nil)
	}
	return v.setValue(key, []byte(strconv.FormatInt(n, 10)))
}

func (v *Vault) setBool(key string, b bool) error {
	s := "0"
	if b {
		s = "1"
	}
	return v.setValue(key, []byte(s))
}

func (v *Vault) setTime(key string, t time.Time) error {
	return v.setInt64(key, tsutil.Millis(t))
}

func (v *Vault) getValue(key string) ([]byte, error) {
	b, err := v.store.Get(key)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (v *Vault) getInt64(key string) (int64, error) {
	b, err := v.getValue(key)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (v *Vault) getTime(key string) (time.Time, error) {
	n, err := v.getInt64(key)
	if err != nil {
		return time.Time{}, err
	}
	if n == 0 {
		return time.Time{}, nil
	}
	return tsutil.ConvertMillis(n), nil
}

func (v *Vault) getBool(key string) (bool, error) {
	b, err := v.getValue(key)
	if err != nil {
		return false, err
	}
	if string(b) == "1" {
		return true, nil
	}
	return false, nil
}

func pad(n int64) string {
	if n > 999999999999999 {
		panic("int too large for padding")
	}
	return fmt.Sprintf("%015d", n)
}

// func unpad(s string) (int64, error) {
// 	n, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return int64(n), nil
// }
