package git

import (
	"strings"

	"github.com/pkg/errors"
)

// ParseHost from git url string.
func ParseHost(urs string) (string, error) {
	spl := strings.SplitN(urs, ":", 2)
	if len(spl) != 2 {
		return "", errors.Errorf("unrecognized git url format")
	}
	hspl := strings.SplitN(spl[0], "@", 2)
	if len(hspl) == 1 {
		return hspl[0], nil
	}
	return hspl[1], nil
}
