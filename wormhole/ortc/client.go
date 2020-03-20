package ortc

import (
	"fmt"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pion/webrtc/v2"
)

type Client struct {
	api      *webrtc.API
	ice      *webrtc.ICETransport
	gatherer *webrtc.ICEGatherer
	dtls     *webrtc.DTLSTransport
	sctp     *webrtc.SCTPTransport
}

func NewClient() (*Client, error) {
	iceOptions := webrtc.ICEGatherOptions{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	api := webrtc.NewAPI()

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

	sctp.OnDataChannel(func(channel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", channel.Label(), channel.ID())

		// Register the handlers
		channel.OnOpen(handleOnOpen(channel))
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", channel.Label(), string(msg.Data))
		})
	})

	return &Client{
		api:      api,
		gatherer: gatherer,
		ice:      ice,
		dtls:     dtls,
		sctp:     sctp,
	}, nil
}

func (c *Client) Close() {
	c.gatherer.Close()
}

func (c *Client) Gather() (*Signal, error) {
	// Gather candidates
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

func (c *Client) SetRemote(signal *Signal, offer bool) error {
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

	// Construct the data channel as the offerer
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

		// Register the handlers
		// channel.OnOpen(handleOnOpen(channel)) // TODO: OnOpen on handle ChannelAck
		go handleOnOpen(channel)() // Temporary alternative
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", channel.Label(), string(msg.Data))
		})
	}

	select {}
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

func handleOnOpen(channel *webrtc.DataChannel) func() {
	return func() {
		fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", channel.Label(), channel.ID())

		for range time.NewTicker(5 * time.Second).C {
			message := keys.RandPhrase()
			fmt.Printf("Sending '%s' \n", message)

			err := channel.SendText(message)
			if err != nil {
				panic(err)
			}
		}
	}
}
