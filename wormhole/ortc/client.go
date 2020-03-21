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

	openLn    func(*webrtc.DataChannel)
	statusLn  func(status Status)
	messageLn func(*webrtc.DataChannel, webrtc.DataChannelMessage)
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
		openLn:    func(*webrtc.DataChannel) {},
		messageLn: func(*webrtc.DataChannel, webrtc.DataChannelMessage) {},
		statusLn:  func(status Status) {},
	}

	ice.OnConnectionStateChange(func(state webrtc.ICETransportState) {
		status := connectionStatus(state)
		logger.Infof("Status: %s", status)
		cl.statusLn(status)
	})

	sctp.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnOpen(func() {
			cl.onOpen(channel)
		})
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			cl.onMessage(channel, msg)
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

	// candidate, err := c.stunCandidate()
	// if err != nil {
	// 	return nil, err
	// }
	// iceCandidates = append(iceCandidates, *candidate)

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

func (c *Client) stunCandidate() (*webrtc.ICECandidate, error) {
	stunAddr, _, err := stunAddress()
	if err != nil {
		return nil, err
	}
	// conn.Close()

	addr := stunAddr.IP.String()
	port := uint16(stunAddr.Port)

	return &webrtc.ICECandidate{
		Foundation:     "foundation",
		Priority:       1694498815,
		Address:        addr,
		Port:           port,
		Protocol:       webrtc.ICEProtocolUDP,
		Typ:            webrtc.ICECandidateTypeSrflx,
		Component:      1,
		RelatedAddress: "0.0.0.0",
		RelatedPort:    port,
	}, nil

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
			c.onOpen(channel)
		})
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			c.onMessage(channel, msg)
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

func (c *Client) onOpen(channel *webrtc.DataChannel) {
	c.openLn(channel)
}

func (c *Client) onMessage(channel *webrtc.DataChannel, msg webrtc.DataChannelMessage) {
	c.messageLn(channel, msg)
}

func (c *Client) OnStatus(f func(Status)) {
	c.statusLn = f
}

func (c *Client) OnOpen(f func(channel *webrtc.DataChannel)) {
	c.openLn = f
}

func (c *Client) OnMessage(f func(channel *webrtc.DataChannel, msg webrtc.DataChannelMessage)) {
	c.messageLn = f
}

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

func connectionStatus(state webrtc.ICETransportState) Status {
	switch state {
	case webrtc.ICETransportStateNew:
		return Initialized
	case webrtc.ICETransportStateChecking:
		return Checking
	case webrtc.ICETransportStateConnected:
		return Connected
	case webrtc.ICETransportStateCompleted:
		return Completed
	case webrtc.ICETransportStateDisconnected:
		return Disconnected
	case webrtc.ICETransportStateFailed:
		return Failed
	case webrtc.ICETransportStateClosed:
		return Closed
	default:
		return Initialized
	}
}
