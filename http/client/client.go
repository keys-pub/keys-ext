package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Client ...
type Client struct {
	url        *url.URL
	httpClient *http.Client
	clock      tsutil.Clock
}

// New creates a Client for an HTTP API.
func New(urs string) (*Client, error) {
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

// ErrResponse ...
type ErrResponse struct {
	StatusCode int
	Message    string
	URL        *url.URL
}

func (e ErrResponse) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Message)
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
	if resp.StatusCode == 200 {
		return nil
	}
	// Default error
	err := ErrResponse{StatusCode: resp.StatusCode, URL: resp.Request.URL}

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

func (c *Client) req(ctx context.Context, method string, path string, params url.Values, key *keys.EdX25519Key, body io.Reader, contentHash string) (*http.Response, error) {
	urs, err := c.urlFor(path, params)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Client req %s %s", method, urs)

	var req *http.Request
	if key != nil {
		r, err := api.NewRequestWithContext(ctx, method, urs, body, contentHash, c.clock.Now(), key)
		if err != nil {
			return nil, err
		}
		req = r
	} else {
		r, err := http.NewRequestWithContext(ctx, method, urs, body)
		if err != nil {
			return nil, err
		}
		req = r
	}

	return c.httpClient.Do(req)
}

func (c *Client) document(path string, resp *http.Response) (*docs.Document, error) {
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

	doc := docs.NewDocument(path, b)
	doc.CreatedAt = createdAt
	doc.UpdatedAt = updatedAt
	return doc, nil
}

func (c *Client) getDocument(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key) (*docs.Document, error) {
	resp, err := c.get(ctx, path, params, key)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.document(path, resp)
}

func (c *Client) get(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key) (*http.Response, error) {
	resp, err := c.req(ctx, "GET", path, params, key, nil, "")
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

func (c *Client) head(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key) (*http.Response, error) {
	resp, err := c.req(ctx, "HEAD", path, params, key, nil, "")
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

// func (c *Client) websocketGet(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key) (*websocket.Conn, error) {
// 	url := c.url.String() + path
// 	query := params.Encode()
// 	if query != "" {
// 		url = url + "?" + query
// 	}

// 	auth, err := api.NewAuth("GET", url, c.clock(), key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch c.url.Scheme {
// 	case "http":
// 		auth.URL.Scheme = "ws"
// 	case "https":
// 		auth.URL.Scheme = "wss"
// 	}

// 	header := http.Header{}
// 	header.Set("Authorization", auth.Header())

// 	logger.Debugf("Websocket dial %s", auth.URL.String())
// 	conn, _, err := websocket.DefaultDialer.Dial(auth.URL.String(), header)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return conn, nil
// }

func (c *Client) put(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key, reader io.Reader, contentHash string) (*http.Response, error) {
	resp, err := c.req(ctx, "PUT", path, params, key, reader, contentHash)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) putDocument(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key, reader io.Reader, contentHash string) (*docs.Document, error) {
	resp, err := c.put(ctx, path, params, key, reader, contentHash)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.document(path, resp)
}

func (c *Client) post(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key, reader io.Reader, contentHash string) (*http.Response, error) {
	resp, err := c.req(ctx, "POST", path, params, key, reader, contentHash)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) postDocument(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key, reader io.Reader, contentHash string) (*docs.Document, error) {
	resp, err := c.post(ctx, path, params, key, reader, contentHash)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	return c.document(path, resp)
}

func (c *Client) delete(ctx context.Context, path string, params url.Values, key *keys.EdX25519Key) (*http.Response, error) {
	resp, err := c.req(ctx, "DELETE", path, params, key, nil, "")
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
