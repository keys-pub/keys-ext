package client

import (
	"encoding/json"
	"io/ioutil"

	"github.com/keys-pub/keys/http"
	"github.com/pkg/errors"
)

func unmarshal(resp *http.Response, v interface{}) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal json response")
	}
	return json.Unmarshal(b, v)
}
