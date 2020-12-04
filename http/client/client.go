package client

import (
	"bytes"
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

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Client ...
type Client struct {
	url        *url.URL
	httpClient *http.Client
	clock      tsutil.Clock
}

// New creates a Client for the keys.pub Web API.
func New(urs string) (*Client, error) {
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
	StatusCode int
	Message    string
	URL        *url.URL
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.StatusCode)
}

// URL ...
func (c *Client) URL() *url.URL {
	return c.url
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
	err := Error{StatusCode: resp.StatusCode, URL: resp.Request.URL}

	b, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return err
	}
	if len(b) == 0 {
		return err
	}
	var respVal api.Response
	if err := json.Unmarshal(b, &respVal); err != nil {
		if utf8.Valid(b) {
			return errors.New(string(b))
		}
		return errors.Errorf("error response not valid utf8")
	}
	if respVal.Error != nil {
		err.Message = respVal.Error.Message
	}

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

func (c *Client) req(ctx context.Context, method string, path string, params url.Values, body io.Reader, contentHash string, auth http.AuthProvider) (*http.Response, error) {
	urs, err := c.urlFor(path, params)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Client req %s %s", method, urs)

	var req *http.Request
	if auth != nil {
		r, err := http.NewAuthRequest(method, urs, body, contentHash, c.clock.Now(), auth)
		if err != nil {
			return nil, err
		}
		req = r
	} else {
		r, err := http.NewRequest(method, urs, body)
		if err != nil {
			return nil, err
		}
		req = r
	}

	return c.httpClient.Do(req.WithContext(ctx))
}

// Response ...
type Response struct {
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Client) response(path string, resp *http.Response) (*Response, error) {
	b, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	createdHeader := resp.Header.Get("CreatedAt-RFC3339M")
	createdAt := time.Time{}
	if createdHeader != "" {
		tm, err := time.Parse(tsutil.RFC3339Milli, createdHeader)
		if err != nil {
			return nil, err
		}
		createdAt = tm
	}

	updatedHeader := resp.Header.Get("Last-Modified-RFC3339M")
	updatedAt := time.Time{}
	if updatedHeader != "" {
		tm, err := time.Parse(tsutil.RFC3339Milli, updatedHeader)
		if err != nil {
			return nil, err
		}
		updatedAt = tm
	}

	out := &Response{Data: b}
	out.CreatedAt = createdAt
	out.UpdatedAt = updatedAt
	return out, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, auth http.AuthProvider) (*Response, error) {
	resp, err := c.req(ctx, "GET", path, params, nil, "", auth)
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
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.response(path, resp)
}

func (c *Client) head(ctx context.Context, path string, params url.Values, auth http.AuthProvider) (*http.Response, error) {
	resp, err := c.req(ctx, "HEAD", path, params, nil, "", auth)
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

func (c *Client) put(ctx context.Context, path string, params url.Values, reader io.Reader, contentHash string, auth http.AuthProvider) (*Response, error) {
	resp, err := c.req(ctx, "PUT", path, params, reader, contentHash, auth)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.response(path, resp)
}

func (c *Client) putRetryOnConflict(ctx context.Context, path string, params url.Values, b []byte, contentHash string, auth http.AuthProvider, attempt int, maxAttempts int, delay time.Duration) (*Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	resp, err := c.put(ctx, path, params, bytes.NewReader(b), contentHash, auth)
	if err != nil {
		var rerr Error
		if attempt < maxAttempts && errors.As(err, &rerr) && rerr.StatusCode == http.StatusConflict {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			return c.putRetryOnConflict(ctx, path, params, b, contentHash, auth, attempt+1, maxAttempts, delay)
		}
		return nil, err
	}
	return resp, nil
}

func (c *Client) post(ctx context.Context, path string, params url.Values, reader io.Reader, contentHash string, auth http.AuthProvider) (*Response, error) {
	resp, err := c.req(ctx, "POST", path, params, reader, contentHash, auth)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.response(path, resp)
}

func (c *Client) delete(ctx context.Context, path string, params url.Values, reader io.Reader, contentHash string, auth http.AuthProvider) (*http.Response, error) {
	resp, err := c.req(ctx, "DELETE", path, params, reader, contentHash, auth)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
