package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pion/webrtc/v2"
)

func main() {
	isOffer := flag.Bool("offer", false, "Act as the offerer if set")
	flag.Parse()

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
		panic(err)
	}

	// Construct the ICE transport
	ice := api.NewICETransport(gatherer)

	// Construct the DTLS transport
	dtls, err := api.NewDTLSTransport(ice, nil)
	if err != nil {
		panic(err)
	}

	// Construct the SCTP transport
	sctp := api.NewSCTPTransport(dtls)

	// Handle incoming data channels
	sctp.OnDataChannel(func(channel *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", channel.Label(), channel.ID())

		// Register the handlers
		channel.OnOpen(handleOnOpen(channel))
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", channel.Label(), string(msg.Data))
		})
	})

	// Gather candidates
	err = gatherer.Gather()
	if err != nil {
		panic(err)
	}

	iceCandidates, err := gatherer.GetLocalCandidates()
	if err != nil {
		panic(err)
	}

	iceParams, err := gatherer.GetLocalParameters()
	if err != nil {
		panic(err)
	}

	dtlsParams, err := dtls.GetLocalParameters()
	if err != nil {
		panic(err)
	}

	sctpCapabilities := sctp.GetCapabilities()

	s := Signal{
		ICECandidates:    iceCandidates,
		ICEParameters:    iceParams,
		DTLSParameters:   dtlsParams,
		SCTPCapabilities: sctpCapabilities,
	}

	// Exchange the information
	fmt.Println(Encode(s))
	remoteSignal := Signal{}
	Decode(MustReadStdin(), &remoteSignal)

	iceRole := webrtc.ICERoleControlled
	if *isOffer {
		iceRole = webrtc.ICERoleControlling
	}

	err = ice.SetRemoteCandidates(remoteSignal.ICECandidates)
	if err != nil {
		panic(err)
	}

	// Start the ICE transport
	err = ice.Start(nil, remoteSignal.ICEParameters, &iceRole)
	if err != nil {
		panic(err)
	}

	// Start the DTLS transport
	err = dtls.Start(remoteSignal.DTLSParameters)
	if err != nil {
		panic(err)
	}

	// Start the SCTP transport
	err = sctp.Start(remoteSignal.SCTPCapabilities)
	if err != nil {
		panic(err)
	}

	// Construct the data channel as the offerer
	if *isOffer {
		var id uint16 = 1

		dcParams := &webrtc.DataChannelParameters{
			Label: "Foo",
			ID:    &id,
		}
		var channel *webrtc.DataChannel
		channel, err = api.NewDataChannel(sctp, dcParams)
		if err != nil {
			panic(err)
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

// Allows compressing offer/answer to bypass terminal input limits.
const compress = false

// MustReadStdin blocks until input is received from stdin
func MustReadStdin() string {
	r := bufio.NewReader(os.Stdin)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				panic(err)
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	fmt.Println("")

	return in
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	if compress {
		b = zip(b)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if compress {
		b = unzip(b)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}

func zip(in []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(in)
	if err != nil {
		panic(err)
	}
	err = gz.Flush()
	if err != nil {
		panic(err)
	}
	err = gz.Close()
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func unzip(in []byte) []byte {
	var b bytes.Buffer
	_, err := b.Write(in)
	if err != nil {
		panic(err)
	}
	r, err := gzip.NewReader(&b)
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return res
}
