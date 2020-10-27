package matter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Client ...
type Client struct {
	url        *url.URL
	httpClient *http.Client
	clock      tsutil.Clock

	AuthType  string
	AuthToken string
}

// NewClient creates a Client for a the mattermost API.
func NewClient(urs string) (*Client, error) {
	urs = strings.TrimSuffix(urs, "/")
	urp, err := url.Parse(urs)
	if err != nil {
		return nil, err
	}

	return &Client{
		url:        urp,
		httpClient: defaultHTTPClient(),
		clock:      tsutil.NewClock(),
	}, nil
}

// URL for client.
func (c *Client) URL() *url.URL {
	return c.url
}

// NewWebSocketClient returns new WebSocketClient.
func (c *Client) NewWebSocketClient() (*WebSocketClient, error) {
	// TODO: Close on logout
	urs := fmt.Sprintf("ws://%s", c.url.Host)
	return NewWebSocketClient(urs, c.AuthToken)
}

// SetHTTPClient sets the http.Client to use.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// TODO: are these timeouts too agressive?
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
}

// Error ...
type Error struct {
	StatusCode int      `json:"status_code,omitempty"`
	Message    string   `json:"message,omitempty"`
	URL        *url.URL `json:"-"`
	ID         string   `json:"id,omitempty"`
	RequestID  string   `json:"request_id,omitempty"`
}

func (e Error) Error() string {
	if e.URL != nil {
		return fmt.Sprintf("%s (%d) %s", e.Message, e.StatusCode, e.URL.String())
	}
	return fmt.Sprintf("%s (%d)", e.Message, e.StatusCode)
}

// SetClock sets the clock Now fn.
func (c *Client) SetClock(clock tsutil.Clock) {
	c.clock = clock
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Default error
	dfltErr := Error{StatusCode: resp.StatusCode, URL: resp.Request.URL}
	b, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return dfltErr
	}
	if len(b) == 0 {
		return dfltErr
	}

	var err Error
	if err := json.Unmarshal(b, &err); err != nil {
		if utf8.Valid(b) {
			dfltErr.Message = string(b)
			return dfltErr
		}
		dfltErr.Message = "error response not valid utf8"
		return dfltErr
	}
	err.URL = resp.Request.URL
	return err
}

func (c *Client) urlFor(path string, params url.Values) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return "", errors.Errorf("req accepts a path, not an url")
	}

	urs := c.url.String() + path
	query := params.Encode()
	if query != "" {
		urs = urs + "?" + query
	}
	return urs, nil
}

func (c *Client) reqWithKey(ctx context.Context, method string, path string, params url.Values, body io.Reader, key *keys.EdX25519Key, contentHash string) (*http.Response, error) {
	urs, err := c.urlFor(path, params)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.Errorf("http request failed: no key specified")
	}

	logger.Debugf("Request %s %s (%s)", method, urs, key.ID())
	req, err := http.NewAuthRequestWithContext(ctx, method, urs, body, contentHash, c.clock.Now(), key)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func (c *Client) req(ctx context.Context, method string, path string, params url.Values, body io.Reader) (*http.Response, error) {
	urs, err := c.urlFor(path, params)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Request %s %s", method, urs)
	req, err := http.NewRequestWithContext(ctx, method, urs, body)
	if err != nil {
		return nil, err
	}

	if len(c.AuthToken) > 0 {
		req.Header.Set("Authorization", c.AuthType+" "+c.AuthToken)
	}

	return c.httpClient.Do(req)
}

// Get request.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	resp, err := c.req(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to GET")
	}
	if resp.StatusCode == 404 {
		logger.Debugf("Not found %s", path)
		return nil, nil
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Head request.
func (c *Client) Head(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	resp, err := c.req(ctx, "HEAD", path, params, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to HEAD")
	}
	if resp.StatusCode == 404 {
		logger.Debugf("Not found %s", path)
		return nil, nil
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Put request.
func (c *Client) Put(ctx context.Context, path string, params url.Values, reader io.Reader) (*http.Response, error) {
	resp, err := c.req(ctx, "PUT", path, params, reader)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Post request.
func (c *Client) Post(ctx context.Context, path string, params url.Values, reader io.Reader) (*http.Response, error) {
	resp, err := c.req(ctx, "POST", path, params, reader)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// PostWithKey request.
func (c *Client) PostWithKey(ctx context.Context, path string, params url.Values, reader io.Reader, key *keys.EdX25519Key, contentHash string) (*http.Response, error) {
	resp, err := c.reqWithKey(ctx, "POST", path, params, reader, key, contentHash)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Delete request.
func (c *Client) Delete(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	resp, err := c.req(ctx, "DELETE", path, params, nil)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
