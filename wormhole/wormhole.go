package wormhole

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/noise"
	httpclient "github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/wormhole/webrtc"
)

type Wormhole struct {
	rtc   *webrtc.Client
	hcl   *httpclient.Client
	noise *noise.Noise
}

// NewWormhole creates a new Wormhole.
func NewWormhole(server string, ks *keys.Keystore) (*Wormhole, error) {
	rtc, err := webrtc.NewClient()
	if err != nil {
		return nil, err
	}

	if server == "" {
		server = "https://keys.pub"
	}

	hcl, err := httpclient.NewClient(server, ks)
	if err != nil {
		return nil, err
	}

	w := &Wormhole{
		rtc: rtc,
		hcl: hcl,
	}
	rtc.OnChannel(w.onChannel)
	rtc.OnMessage(w.onMessage)

	return w, nil
}

func (w *Wormhole) onMessage(message *webrtc.DataChannelMessage) {

}

func (w *Wormhole) onChannel(message *webrtc.DataChannel) {

}

func (w *Wormhole) Offer(sender *keys.X25519Key, recipient *keys.X25519PublicKey) error {
	offer, err := w.rtc.Offer("wormhole")
	if err != nil {
		return err
	}

	noise, err := noise.NewNoise(sender, recipient, true)
	if err != nil {
		return err
	}

}

func (w *Wormhole) Answer(sender *keys.X25519Key, recipient *keys.X25519PublicKey) error {
	offer, err := w.client.Offer("wormhole")
	if err != nil {
		return err
	}

	noise, err := noise.NewNoise(sender, recipient, true)
	if err != nil {
		return err
	}

}
