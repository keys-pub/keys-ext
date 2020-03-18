package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

type DataChannel = webrtc.DataChannel
type DataChannelMessage = webrtc.DataChannelMessage
type SessionDescription = webrtc.SessionDescription

// Client for webrtc.
type Client struct {
	sync.Mutex
	config  webrtc.Configuration
	conn    *webrtc.PeerConnection
	channel *DataChannel

	channelLn func(msg *DataChannel)
	messageLn func(msg *DataChannelMessage)
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
		channelLn: func(msg *DataChannel) {},
		messageLn: func(msg *DataChannelMessage) {},
	}

	return c, nil
}

func (c *Client) newConnection() (*webrtc.PeerConnection, error) {
	conn, err := webrtc.NewPeerConnection(c.config)
	if err != nil {
		return nil, err
	}

	conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logger.Infof("ICE: %s", connectionState.String())
	})

	conn.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			c.onChannel(d)
		})
		d.OnMessage(func(m DataChannelMessage) {
			c.onMessage(&m)
		})
	})

	return conn, nil
}

func (c *Client) Offer(label string) (*webrtc.SessionDescription, error) {
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

	channel, err := conn.CreateDataChannel(label, nil)
	if err != nil {
		return nil, err
	}
	channel.OnOpen(func() {
		c.onChannel(channel)
	})
	channel.OnMessage(func(m DataChannelMessage) {
		c.onMessage(&m)
	})

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

func (c *Client) Channel() *DataChannel {
	return c.channel
}

func (c *Client) onChannel(channel *DataChannel) {
	c.channel = channel
	c.channelLn(channel)
}

func (c *Client) onMessage(msg *DataChannelMessage) {
	c.messageLn(msg)
}

func (c *Client) OnChannel(f func(msg *DataChannel)) {
	c.channelLn = f
}

func (c *Client) OnMessage(f func(msg *DataChannelMessage)) {
	c.messageLn = f
}

func (c *Client) Send(data []byte) error {
	if c.channel == nil {
		return errors.Errorf("no channel")
	}
	return c.channel.Send(data)
}
