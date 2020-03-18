package wormhole

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/noise"
	httpclient "github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/wormhole/webrtc"
	"github.com/pkg/errors"
)

// ErrOfferNotFound error if offer not found for recipient.
var ErrOfferNotFound = errors.New("offer not found")

// Wormhole for connecting two participants using webrtc, noise and
// keys.pub.
type Wormhole struct {
	sync.Mutex
	rtc   *webrtc.Client
	hcl   *httpclient.Client
	noise *noise.Noise

	onChannel func(label string, id uint16)
	onMessage func(b []byte)

	offerDelay time.Duration
}

// NewWormhole creates a new Wormhole.
// Server is offer/answer message server, only used to coordinate starting the
// webrtc channel.
func NewWormhole(server string, ks *keys.Keystore) (*Wormhole, error) {
	rtc, err := webrtc.NewClient()
	if err != nil {
		return nil, err
	}

	if server == "" {
		server = "https://keys.pub"
	}

	logger.Infof("New wormhole: %s", server)
	hcl, err := httpclient.NewClient(server, ks)
	if err != nil {
		return nil, err
	}

	w := &Wormhole{
		rtc:        rtc,
		hcl:        hcl,
		onChannel:  func(label string, id uint16) {},
		onMessage:  func(b []byte) {},
		offerDelay: time.Second * 2,
	}

	return w, nil
}

// Close wormhole.
func (w *Wormhole) Close() {
	w.rtc.Close()
}

// SetTimeNow sets wormhole clock.
func (w *Wormhole) SetTimeNow(nowFn func() time.Time) {
	w.hcl.SetTimeNow(nowFn)
}

func (w *Wormhole) OnChannel(f func(label string, id uint16)) {
	w.onChannel = f
}

func (w *Wormhole) OnMessage(f func(b []byte)) {
	w.onMessage = f
}

func (w *Wormhole) messageLn(message *webrtc.DataChannelMessage) {
	logger.Infof("Message (%d)", len(message.Data))
	decrypted, err := w.noise.Decrypt(nil, nil, message.Data)
	if err != nil {
		logger.Errorf("Failed to decrypt message: %s", err)
		return
	}
	w.onMessage(decrypted)
}

func (w *Wormhole) channelLn(channel *webrtc.DataChannel) {
	logger.Infof("Channel: %+v", channel)
	id := uint16(0)
	if channel.ID() != nil {
		id = *channel.ID()
	}
	w.onChannel(channel.Label(), id)
}

func (w *Wormhole) Start(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey) error {
	w.Lock()
	defer w.Unlock()

	if w.noise != nil {
		return errors.Errorf("wormhole already started")
	}
	// Initiator is whichever ID is less than
	initiator := sender.ID() < recipient.ID()

	noise, err := noise.NewNoise(sender.X25519Key(), recipient.X25519PublicKey(), initiator)
	if err != nil {
		return err
	}
	w.noise = noise

	var noiseErr error
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Start handshake when channel is connected.
	w.rtc.OnChannel(func(channel *webrtc.DataChannel) {
		if initiator {
			logger.Infof("Initiate handshake...")
			if err := w.handshakeWrite(); err != nil {
				noiseErr = err
				wg.Done()
				return
			}
		}
	})

	w.rtc.OnMessage(func(message *webrtc.DataChannelMessage) {
		logger.Infof("Handshake received...")
		if _, err := noise.HandshakeRead(message.Data); err != nil {
			noiseErr = err
			wg.Done()
			return
		}
		if !initiator {
			logger.Infof("Handshake respond...")
			if err := w.handshakeWrite(); err != nil {
				noiseErr = err
				wg.Done()
				return
			}
			wg.Done()
		} else {
			wg.Done()
		}
	})

	if initiator {
		if err := w.offer(ctx, sender, recipient, w.offerDelay); err != nil {
			return err
		}
	} else {
		if err := w.answer(ctx, sender, recipient, w.offerDelay); err != nil {
			return err
		}
	}

	c := make(chan bool)
	go func() {
		logger.Infof("Wait for handshake...")
		wg.Wait()
		c <- true
		logger.Infof("Handshake wait over")
	}()

	select {
	case <-c:
		break
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 5):
		return errors.Errorf("handshake timeout")
	}

	if noiseErr != nil {
		logger.Errorf("Handshake error: %v", noiseErr)
		return noiseErr
	}

	logger.Infof("Started")
	w.rtc.OnMessage(w.messageLn)
	w.channelLn(w.rtc.Channel())

	return nil
}

// Send data.
func (w *Wormhole) Send(b []byte) error {
	if w.noise == nil {
		return errors.Errorf("no channel (noise)")
	}
	encrypted, err := w.noise.Encrypt(nil, nil, b)
	if err != nil {
		return err
	}
	return w.rtc.Send(encrypted)
}

func (w *Wormhole) offer(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey, delay time.Duration) error {
	logger.Infof("Creating webrtc offer...")
	offer, err := w.rtc.Offer("wormhole")
	if err != nil {
		return err
	}
	b, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	logger.Infof("Offer: %s", string(b))

	logger.Infof("Sending offer message...")
	opts := &httpclient.MessageOpts{
		Channel: "wormhole",
	}
	msg, err := w.hcl.SendMessage(sender, recipient.ID(), b, opts)
	if err != nil {
		return err
	}
	logger.Infof("Offer sent")

	logger.Infof("Wait for answer...")
	answer, err := w.findSession(ctx, sender, recipient, msg.ID, delay)
	if err != nil {
		return err
	}
	if answer == nil {
		return ErrOfferNotFound
	}

	logger.Infof("Setting answer...")
	if err := w.rtc.SetAnswer(answer); err != nil {
		return err
	}

	return nil
}

func (w *Wormhole) findSession(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey, sessionMsgID string, delay time.Duration) (*webrtc.SessionDescription, error) {
	offer, err := w.findMessage(sender, recipient, sessionMsgID)
	if err != nil {
		return nil, err
	}
	if offer != nil {
		return offer, nil
	}
	select {
	case <-time.After(delay):
		return w.findSession(ctx, sender, recipient, sessionMsgID, delay)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (w *Wormhole) findMessage(sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey, sessionMsgID string) (*webrtc.SessionDescription, error) {
	logger.Infof("Getting wormhole messages...")
	opts := &httpclient.MessagesOpts{
		Channel:   "wormhole",
		Limit:     1,
		Direction: keys.Descending,
	}
	msgs, _, err := w.hcl.Messages(sender, recipient.ID(), opts)
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		if msg.ID == sessionMsgID {
			continue
		}
		logger.Infof("Found message, decrypting...")
		b, pk, err := w.hcl.DecryptMessage(sender, msgs[0])
		if err != nil {
			return nil, err
		}
		if pk != recipient.ID() {
			return nil, errors.Errorf("offer not by recipient %s != %s", pk, recipient.ID())
		}

		var offer webrtc.SessionDescription
		if err := json.Unmarshal(b, &offer); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal wormhole offer")
		}
		logger.Infof("Found offer: %s", string(b))
		return &offer, nil
	}
	return nil, nil
}

func (w *Wormhole) answer(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey, delay time.Duration) error {
	logger.Infof("Find offer...")
	offer, err := w.findSession(ctx, sender, recipient, "", delay)
	if err != nil {
		return err
	}

	if offer == nil {
		return ErrOfferNotFound
	}

	logger.Infof("Creating answer...")
	answer, err := w.rtc.Answer(offer)
	if err != nil {
		return err
	}

	b, err := json.Marshal(answer)
	if err != nil {
		return err
	}
	logger.Infof("Answer: %s", string(b))

	logger.Infof("Sending answer message...")
	opts := &httpclient.MessageOpts{
		Channel: "wormhole",
	}
	if _, err := w.hcl.SendMessage(sender, recipient.ID(), b, opts); err != nil {
		return err
	}
	logger.Infof("Answer sent")

	return nil
}

func (w *Wormhole) handshakeWrite() error {
	out, err := w.noise.HandshakeWrite(nil)
	if err != nil {
		return err
	}
	if err := w.rtc.Send(out); err != nil {
		return err
	}
	return nil
}
