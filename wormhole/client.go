package wormhole

import (
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

type Client struct {
	conn    *webrtc.PeerConnection
	channel *webrtc.DataChannel

	channelLn func(msg *webrtc.DataChannel)
	messageLn func(msg webrtc.DataChannelMessage)
}

func NewClient() (*Client, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection.
	conn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	// Set the handler for ICE connection state.
	// This will notify you when the peer has connected/disconnected.
	conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logger.Infof("ICE Connection State has changed: %s\n", connectionState.String())
	})

	c := &Client{
		conn:      conn,
		channelLn: func(msg *webrtc.DataChannel) {},
		messageLn: func(msg webrtc.DataChannelMessage) {},
	}

	// OnDataChannel (answer)
	conn.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			c.onChannel(d)
		})
		d.OnMessage(c.onMessage)
	})

	return c, nil
}

func (c *Client) Offer() (*webrtc.SessionDescription, error) {
	// Create a datachannel with label 'wormhole'
	dataChannel, err := c.conn.CreateDataChannel("wormhole", nil)
	if err != nil {
		return nil, err
	}

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		c.onChannel(dataChannel)
	})

	// Register text message handling
	dataChannel.OnMessage(c.onMessage)

	// Create an offer to send to the browser
	offer, err := c.conn.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := c.conn.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	return &offer, nil
}

func (c *Client) Answer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if err := c.conn.SetRemoteDescription(*offer); err != nil {
		return nil, err
	}

	// Create answer
	answer, err := c.conn.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := c.conn.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return &answer, nil
}

func (c *Client) SetAnswer(answer *webrtc.SessionDescription) error {
	if err := c.conn.SetRemoteDescription(*answer); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
}

func (c *Client) onChannel(channel *webrtc.DataChannel) {
	c.channel = channel
	c.channelLn(channel)
}

func (c *Client) onMessage(msg webrtc.DataChannelMessage) {
	c.messageLn(msg)
}

func (c *Client) OnChannel(f func(msg *webrtc.DataChannel)) {
	c.channelLn = f
}

func (c *Client) OnMessage(f func(msg webrtc.DataChannelMessage)) {
	c.messageLn = f
}

func (c *Client) Send(data []byte) error {
	if c.channel == nil {
		return errors.Errorf("no channel")
	}
	return c.channel.Send(data)
}
