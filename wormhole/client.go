package wormhole

import (
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

type Client struct {
	api    *webrtc.API
	ice    *webrtc.ICETransport
	dtls   *webrtc.DTLSTransport
	sctp   *webrtc.SCTPTransport
	signal *Signal

	channel *webrtc.DataChannel

	onChannel func(msg *webrtc.DataChannel)
	onMessage func(msg webrtc.DataChannelMessage)
}

func NewClient() (*Client, error) {
	// Prepare ICE gathering options
	iceOptions := webrtc.ICEGatherOptions{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	// Create an API object
	api := webrtc.NewAPI()

	// Create the ICE gatherer
	gatherer, err := api.NewICEGatherer(iceOptions)
	if err != nil {
		return nil, err
	}

	// Construct the ICE transport
	ice := api.NewICETransport(gatherer)

	// Construct the DTLS transport
	dtls, err := api.NewDTLSTransport(ice, nil)
	if err != nil {
		return nil, err
	}

	// Construct the SCTP transport
	sctp := api.NewSCTPTransport(dtls)

	// Gather candidates
	logger.Infof("Gather candidates...")
	if err := gatherer.Gather(); err != nil {
		return nil, err
	}

	iceCandidates, err := gatherer.GetLocalCandidates()
	if err != nil {
		return nil, err
	}

	iceParams, err := gatherer.GetLocalParameters()
	if err != nil {
		return nil, err
	}

	dtlsParams, err := dtls.GetLocalParameters()
	if err != nil {
		return nil, err
	}

	sctpCapabilities := sctp.GetCapabilities()

	signal := &Signal{
		ICECandidates:    iceCandidates,
		ICEParameters:    iceParams,
		DTLSParameters:   dtlsParams,
		SCTPCapabilities: sctpCapabilities,
	}

	logger.Infof("Signal: %+v", signal)

	c := &Client{
		api:       api,
		ice:       ice,
		dtls:      dtls,
		sctp:      sctp,
		signal:    signal,
		onChannel: func(msg *webrtc.DataChannel) {},
		onMessage: func(msg webrtc.DataChannelMessage) {},
	}

	c.sctp.OnDataChannel(c.setChannel)

	return c, nil
}

func (c *Client) Close() {
	// TODO
}

func (c *Client) Signal() *Signal {
	return c.signal
}

func (c *Client) OnChannel(f func(msg *webrtc.DataChannel)) {
	c.onChannel = f
}

func (c *Client) OnMessage(f func(msg webrtc.DataChannelMessage)) {
	c.onMessage = f
}

func (c *Client) Send(data []byte) error {
	if c.channel == nil {
		return errors.Errorf("no open channel")
	}
	if err := c.channel.Send(data); err != nil {
		return err
	}
	return nil
}

// Start a channel.
// If offer is true, create the channel.
func (c *Client) Start(remote *Signal, offer bool) error {
	iceRole := webrtc.ICERoleControlled
	if offer {
		iceRole = webrtc.ICERoleControlling
	}

	if err := c.ice.SetRemoteCandidates(remote.ICECandidates); err != nil {
		return err
	}

	logger.Infof("Start ICE (%s)...", iceRole)
	// Start the ICE transport
	if err := c.ice.Start(nil, remote.ICEParameters, &iceRole); err != nil {
		return err
	}

	logger.Infof("Start DTLS...")
	// Start the DTLS transport
	if err := c.dtls.Start(remote.DTLSParameters); err != nil {
		return err
	}

	logger.Infof("Start SCTP...")
	// Start the SCTP transport
	if err := c.sctp.Start(remote.SCTPCapabilities); err != nil {
		return err
	}

	// Construct the data channel as the offerer
	if offer {
		var id uint16 = 1

		dcParams := &webrtc.DataChannelParameters{
			Label: "wormhole",
			ID:    &id,
		}
		var channel *webrtc.DataChannel
		channel, err := c.api.NewDataChannel(c.sctp, dcParams)
		if err != nil {
			return err
		}

		c.setChannel(channel)
	}
	return nil
}

func (c *Client) setChannel(channel *webrtc.DataChannel) {
	logger.Infof("New channel %s %d", channel.Label(), channel.ID())
	channel.OnOpen(func() {
		c.channel = channel
		c.onChannel(c.channel)
	})
	channel.OnMessage(c.onMessage)
}

// Signal is used to exchange signaling info.
// This is not part of the ORTC spec. You are free
// to exchange this information any way you want.
type Signal struct {
	ICECandidates    []webrtc.ICECandidate   `json:"iceCandidates"`
	ICEParameters    webrtc.ICEParameters    `json:"iceParameters"`
	DTLSParameters   webrtc.DTLSParameters   `json:"dtlsParameters"`
	SCTPCapabilities webrtc.SCTPCapabilities `json:"sctpCapabilities"`
}
