package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pion/quic"
	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"
)

const messageSize = 15

func main() {
	isOffer := flag.Bool("offer", false, "Act as the offerer if set")
	flag.Parse()

	// This example shows off the experimental implementation of webrtc-quic.

	// Everything below is the Pion WebRTC (ORTC) API! Thanks for using it ❤️.

	// Create an API object
	api := webrtc.NewAPI()

	// Prepare ICE gathering options
	iceOptions := webrtc.ICEGatherOptions{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	// Create the ICE gatherer
	gatherer, err := api.NewICEGatherer(iceOptions)
	if err != nil {
		panic(err)
	}

	// Construct the ICE transport
	ice := api.NewICETransport(gatherer)

	// Construct the Quic transport
	qt, err := api.NewQUICTransport(ice, nil)
	if err != nil {
		panic(err)
	}

	// Handle incoming streams
	qt.OnBidirectionalStream(func(stream *quic.BidirectionalStream) {
		fmt.Printf("New stream %d\n", stream.StreamID())

		// Handle reading from the stream
		go ReadLoop(stream)

		// Handle writing to the stream
		go WriteLoop(stream)
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

	quicParams, err := qt.GetLocalParameters()
	if err != nil {
		panic(err)
	}

	s := &Signal{
		ICECandidates:  iceCandidates,
		ICEParameters:  iceParams,
		QuicParameters: quicParams,
	}

	fmt.Printf("Signal:\n")
	if err := writeSession(s); err != nil {
		panic(err)
	}

	fmt.Printf("Enter remote:\n")
	remoteSignal, err := readSession()
	if err != nil {
		panic(err)
	}

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

	// Start the Quic transport
	err = qt.Start(remoteSignal.QuicParameters)
	if err != nil {
		panic(err)
	}

	// Construct the stream as the offerer
	if *isOffer {
		var stream *quic.BidirectionalStream
		stream, err = qt.CreateBidirectionalStream()
		if err != nil {
			panic(err)
		}

		// Handle reading from the stream
		go ReadLoop(stream)

		// Handle writing to the stream
		go WriteLoop(stream)
	}

	select {}
}

// Signal is used to exchange signaling info.
// This is not part of the ORTC spec. You are free
// to exchange this information any way you want.
type Signal struct {
	ICECandidates  []webrtc.ICECandidate `json:"iceCandidates"`
	ICEParameters  webrtc.ICEParameters  `json:"iceParameters"`
	QuicParameters webrtc.QUICParameters `json:"quicParameters"`
}

// ReadLoop reads from the stream
func ReadLoop(s *quic.BidirectionalStream) {
	for {
		buffer := make([]byte, messageSize)
		params, err := s.ReadInto(buffer)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Message from stream '%d': %s\n", s.StreamID(), string(buffer[:params.Amount]))
	}
}

// WriteLoop writes to the stream
func WriteLoop(s *quic.BidirectionalStream) {
	for range time.NewTicker(5 * time.Second).C {
		message := keys.RandPhrase()
		fmt.Printf("Sending %s \n", message)

		data := quic.StreamWriteParameters{
			Data: []byte(message),
		}
		err := s.Write(data)
		if err != nil {
			panic(err)
		}
	}
}

func readSession() (*Signal, error) {
	scanner := bufio.NewScanner(os.Stdin)
	input := ""

	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			input = input + strings.TrimSpace(text)
		} else {
			dec, err := encoding.Decode(input, encoding.Base64)
			if err != nil {
				return nil, err
			}

			r, err := gzip.NewReader(bytes.NewBuffer(dec))
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				return nil, err
			}

			var signal Signal
			if err := json.Unmarshal(buf.Bytes(), &signal); err != nil {
				log.Fatal(err)
			}
			return &signal, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.Errorf("no input")
}

func writeSession(s *Signal) error {
	mb, err := json.Marshal(s)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(mb); err != nil {
		return err
	}
	gz.Flush()
	gz.Close()
	enc, err := encoding.Encode(b.Bytes(), encoding.Base64)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n\n", enc)
	return nil
}
