package ortc

import (
	"os"

	"github.com/pion/logging"
	"github.com/pion/webrtc/v2"
)

type Client struct {
	api      *webrtc.API
	ice      *webrtc.ICETransport
	gatherer *webrtc.ICEGatherer
	dtls     *webrtc.DTLSTransport
	sctp     *webrtc.SCTPTransport

	onOpen    func(*webrtc.DataChannel)
	onMessage func(*webrtc.DataChannel, webrtc.DataChannelMessage)
}

func NewClient() (*Client, error) {
	iceOptions := webrtc.ICEGatherOptions{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	api, err := newAPI(false)
	if err != nil {
		return nil, err
	}

	gatherer, err := api.NewICEGatherer(iceOptions)
	if err != nil {
		return nil, err
	}

	ice := api.NewICETransport(gatherer)

	dtls, err := api.NewDTLSTransport(ice, nil)
	if err != nil {
		return nil, err
	}

	sctp := api.NewSCTPTransport(dtls)

	cl := &Client{
		api:       api,
		gatherer:  gatherer,
		ice:       ice,
		dtls:      dtls,
		sctp:      sctp,
		onOpen:    func(*webrtc.DataChannel) {},
		onMessage: func(*webrtc.DataChannel, webrtc.DataChannelMessage) {},
	}

	sctp.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnOpen(func() {
			cl.openLn(channel)
		})
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			cl.messageLn(channel, msg)
		})
	})

	return cl, nil
}

func (c *Client) Close() {
	c.gatherer.Close()
}

func newAPI(trace bool) (*webrtc.API, error) {
	wlg := logging.NewDefaultLoggerFactory()
	if trace {
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

func (c *Client) Gather() (*Signal, error) {
	if err := c.gatherer.Gather(); err != nil {
		return nil, err
	}

	iceCandidates, err := c.gatherer.GetLocalCandidates()
	if err != nil {
		return nil, err
	}

	iceParams, err := c.gatherer.GetLocalParameters()
	if err != nil {
		return nil, err
	}

	dtlsParams, err := c.dtls.GetLocalParameters()
	if err != nil {
		return nil, err
	}

	sctpCapabilities := c.sctp.GetCapabilities()

	s := &Signal{
		ICECandidates:    iceCandidates,
		ICEParameters:    iceParams,
		DTLSParameters:   dtlsParams,
		SCTPCapabilities: sctpCapabilities,
	}
	return s, nil
}

func (c *Client) Start(signal *Signal, offer bool) error {
	iceRole := webrtc.ICERoleControlled
	if offer {
		iceRole = webrtc.ICERoleControlling
	}

	if err := c.ice.SetRemoteCandidates(signal.ICECandidates); err != nil {
		return err
	}

	if err := c.ice.Start(nil, signal.ICEParameters, &iceRole); err != nil {
		return err
	}

	if err := c.dtls.Start(signal.DTLSParameters); err != nil {
		return err
	}

	if err := c.sctp.Start(signal.SCTPCapabilities); err != nil {
		return err
	}

	if offer {
		var id uint16 = 1

		dcParams := &webrtc.DataChannelParameters{
			Label: "testing",
			ID:    &id,
		}
		channel, err := c.api.NewDataChannel(c.sctp, dcParams)
		if err != nil {
			return err
		}

		channel.OnOpen(func() {
			c.openLn(channel)
		})
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			c.messageLn(channel, msg)
		})
	}

	return nil
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

func (c *Client) openLn(channel *webrtc.DataChannel) {
	c.onOpen(channel)
}

func (c *Client) messageLn(channel *webrtc.DataChannel, msg webrtc.DataChannelMessage) {
	c.onMessage(channel, msg)
}

func (c *Client) OnOpen(f func(channel *webrtc.DataChannel)) {
	c.onOpen = f
}

func (c *Client) OnMessage(f func(channel *webrtc.DataChannel, msg webrtc.DataChannelMessage)) {
	c.onMessage = f
}
