package webrtc

import (
	"os"
	"sync"

	"github.com/keys-pub/keys"

	"github.com/pion/logging"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

// Channel.
type Channel interface {
	Label() string
	Send(b []byte) error
	OnClose(f func())
}

// Message in channel.
type Message interface {
	Data() []byte
}

type message struct {
	webrtc.DataChannelMessage
}

func (m message) Data() []byte {
	return m.DataChannelMessage.Data
}

type SessionDescription = webrtc.SessionDescription

type Status string

const (
	Initialized  Status = "init"
	Checking     Status = "checking"
	Connected    Status = "connected"
	Completed    Status = "completed"
	Disconnected Status = "disconnected"
	Failed       Status = "failed"
	Closed       Status = "closed"
)

// Client for webrtc.
type Client struct {
	sync.Mutex
	config  webrtc.Configuration
	conn    *webrtc.PeerConnection
	channel *webrtc.DataChannel
	trace   bool

	openLn    func(channel Channel)
	closeLn   func(channel Channel)
	statusLn  func(status Status)
	messageLn func(channel Channel, msg Message)
}

// NewClient creates webrtc Client.
func NewClient() (*Client, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	c := &Client{
		config:    config,
		openLn:    func(channel Channel) {},
		closeLn:   func(channel Channel) {},
		messageLn: func(channel Channel, msg Message) {},
		statusLn:  func(status Status) {},
	}

	return c, nil
}

func (c *Client) newAPI() (*webrtc.API, error) {
	wlg := logging.NewDefaultLoggerFactory()
	if c.trace {
		wlg.DefaultLogLevel = logging.LogLevelTrace
	}
	// wlg.DefaultLogLevel = logging.LogLevelDebug
	wlg.Writer = os.Stderr
	se := webrtc.SettingEngine{
		LoggerFactory: wlg,
	}
	api := webrtc.NewAPI(webrtc.WithSettingEngine(se))
	return api, nil
}

func (c *Client) SetTrace(trace bool) {
	c.trace = trace
}

func (c *Client) newConnection() (*webrtc.PeerConnection, error) {
	api, err := c.newAPI()
	if err != nil {
		return nil, err
	}
	conn, err := api.NewPeerConnection(c.config)
	if err != nil {
		return nil, err
	}

	conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		status := connectionStatus(connectionState)
		logger.Infof("Status: %s", status)
		c.statusLn(status)
	})

	conn.OnDataChannel(func(channel *webrtc.DataChannel) {
		logger.Infof("Data channel: %s", channel.Label())
		c.register(channel)
	})

	return conn, nil
}

func (c *Client) register(channel *webrtc.DataChannel) {
	channel.OnOpen(func() {
		c.onOpen(channel)
	})
	channel.OnClose(func() {
		c.onClose(channel)
	})
	channel.OnMessage(func(m webrtc.DataChannelMessage) {
		c.onMessage(channel, m)
	})
}

func (c *Client) Offer() (*webrtc.SessionDescription, error) {
	c.Lock()
	defer c.Unlock()

	if c.conn != nil {
		return nil, errors.Errorf("connection already exists")
	}
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}
	c.conn = conn

	channel, err := c.conn.CreateDataChannel(keys.RandIDString(), nil)
	if err != nil {
		return nil, err
	}
	c.register(channel)

	offer, err := conn.CreateOffer(nil)
	if err != nil {
		return nil, err
	}
	if err := conn.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return &offer, nil
}

func (c *Client) Answer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	c.Lock()
	defer c.Unlock()

	if c.conn != nil {
		return nil, errors.Errorf("connection already exists")
	}
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}
	c.conn = conn

	if err := conn.SetRemoteDescription(*offer); err != nil {
		return nil, err
	}

	answer, err := conn.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := conn.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return &answer, nil
}

func (c *Client) SetAnswer(answer *webrtc.SessionDescription) error {
	c.Lock()
	defer c.Unlock()

	if c.conn == nil {
		return errors.Errorf("no connection")
	}

	if err := c.conn.SetRemoteDescription(*answer); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			logger.Warningf("Error closing webrtc channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.Warningf("Error closing webrtc connection: %v", err)
		}
	}
}

func (c *Client) Channel() Channel {
	return c.channel
}

func (c *Client) onOpen(channel *webrtc.DataChannel) {
	c.channel = channel
	c.openLn(channel)
}

func (c *Client) onClose(channel *webrtc.DataChannel) {
	c.channel = nil
	c.closeLn(channel)
}

func (c *Client) onMessage(channel *webrtc.DataChannel, m webrtc.DataChannelMessage) {
	c.messageLn(channel, &message{m})
}

func (c *Client) OnStatus(f func(Status)) {
	c.statusLn = f
}

func (c *Client) OnOpen(f func(Channel)) {
	c.openLn = f
}

func (c *Client) OnClose(f func(Channel)) {
	c.closeLn = f
}

func (c *Client) OnMessage(f func(Channel, Message)) {
	c.messageLn = f
}

func (c *Client) Send(data []byte) error {
	if c.channel == nil {
		return errors.Errorf("no channel")
	}
	return c.channel.Send(data)
}

func connectionStatus(connectionState webrtc.ICEConnectionState) Status {
	switch connectionState {
	case webrtc.ICEConnectionStateNew:
		return Initialized
	case webrtc.ICEConnectionStateChecking:
		return Checking
	case webrtc.ICEConnectionStateConnected:
		return Connected
	case webrtc.ICEConnectionStateCompleted:
		return Completed
	case webrtc.ICEConnectionStateDisconnected:
		return Disconnected
	case webrtc.ICEConnectionStateFailed:
		return Failed
	case webrtc.ICEConnectionStateClosed:
		return Closed
	default:
		return Initialized
	}
}
