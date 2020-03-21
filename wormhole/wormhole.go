package wormhole

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/noise"
	httpclient "github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/pkg/errors"
)

// ErrNoResponse error if offer not found for recipient.
var ErrNoResponse = errors.New("no response")

// ErrNoiseHandshakeTimeout if we timed out during handshake
var ErrNoiseHandshakeTimeout = errors.New("noise handshake timeout")

// Wormhole for connecting two participants using webrtc, noise and
// keys.pub.
type Wormhole struct {
	sync.Mutex
	rtc   *sctp.Client
	hcl   *httpclient.Client
	noise *noise.Noise

	onConnect func()
	onClose   func()
}

// NewWormhole creates a new Wormhole.
// Server is offer/answer message server, only used to coordinate starting the
// webrtc channel.
func NewWormhole(server string, ks *keys.Keystore) (*Wormhole, error) {
	rtc := sctp.NewClient()

	if server == "" {
		server = "https://keys.pub"
	}

	logger.Infof("New wormhole (%s)", server)
	hcl, err := httpclient.NewClient(server, ks)
	if err != nil {
		return nil, err
	}

	w := &Wormhole{
		rtc:       rtc,
		hcl:       hcl,
		onConnect: func() {},
		onClose:   func() {},
	}

	// TODO: Close
	// rtc.OnClose(func() {
	// 	w.onClose()
	// })

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

func (w *Wormhole) OnConnect(f func()) {
	w.onConnect = f
}

func (w *Wormhole) OnClose(f func()) {
	w.onClose = f
}

func (w *Wormhole) Start(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey) error {
	w.Lock()
	defer w.Unlock()

	if w.noise != nil {
		return errors.Errorf("wormhole already started")
	}
	// Initiator is whichever ID is less than
	initiator := sender.ID() < recipient.ID()
	logger.Infof("Wormhole initator: %t", initiator)

	if initiator {
		if err := w.connect(ctx, sender.ID(), recipient.ID()); err != nil {
			return err
		}
	} else {
		if err := w.listen(ctx, sender.ID(), recipient.ID()); err != nil {
			return err
		}
	}

	noise, err := noise.NewNoise(sender.X25519Key(), recipient.X25519PublicKey(), initiator)
	if err != nil {
		return err
	}

	// TODO: Noise handshake timeout
	if initiator {
		out, err := noise.HandshakeWrite(nil)
		if err != nil {
			return err
		}
		if err := w.rtc.Write(ctx, out); err != nil {
			return err
		}
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(ctx, buf)
		if err != nil {
			return err
		}
		if _, err := noise.HandshakeRead(buf[:n]); err != nil {
			return err
		}
	} else {
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(ctx, buf)
		if err != nil {
			return err
		}
		if _, err := noise.HandshakeRead(buf[:n]); err != nil {
			return err
		}
		out, err := noise.HandshakeWrite(nil)
		if err != nil {
			return err
		}
		if err := w.rtc.Write(ctx, out); err != nil {
			return err
		}
	}
	w.noise = noise

	logger.Infof("Started")
	w.onConnect()

	return nil
}

// Send data.
func (w *Wormhole) Send(ctx context.Context, b []byte) error {
	if w.noise == nil {
		return errors.Errorf("no channel (noise)")
	}
	encrypted, err := w.noise.Encrypt(nil, nil, b)
	if err != nil {
		return err
	}
	return w.rtc.Write(ctx, encrypted)
}

// Read.
func (w *Wormhole) Read(ctx context.Context) ([]byte, error) {
	buf := make([]byte, 1024)

	n, err := w.rtc.Read(ctx, buf)
	if err != nil {
		return nil, err
	}

	logger.Infof("Wormhole read (%d)", n)
	decrypted, err := w.noise.Decrypt(nil, nil, buf[:n])
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func (w *Wormhole) connect(ctx context.Context, sender keys.ID, recipient keys.ID) error {
	logger.Infof("Wormhole connect...")
	offer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return err
	}
	if err := w.writeOffer(ctx, offer, sender, recipient); err != nil {
		return err
	}

	answer, err := w.readAnswer(ctx, sender, recipient)
	if err != nil {
		return err
	}
	if answer == nil {
		return ErrNoResponse
	}

	if err := w.rtc.Connect(ctx, answer); err != nil {
		return err
	}

	return nil
}

func (w *Wormhole) listen(ctx context.Context, sender keys.ID, recipient keys.ID) error {
	logger.Infof("Wormhole listen...")
	answer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return err
	}
	if err := w.writeAnswer(ctx, answer, sender, recipient); err != nil {
		return err
	}

	offer, err := w.readOffer(ctx, sender, recipient)
	if err != nil {
		return err
	}
	if offer == nil {
		return ErrNoResponse
	}

	if err := w.rtc.Listen(context.TODO(), offer); err != nil {
		return err
	}

	return nil
}

func (w *Wormhole) writeOffer(ctx context.Context, offer *sctp.Addr, sender keys.ID, recipient keys.ID) error {
	return w.writeSession(ctx, offer, sender, recipient, "offer")
}

func (w *Wormhole) readOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient, "offer")
}

func (w *Wormhole) writeAnswer(ctx context.Context, answer *sctp.Addr, sender keys.ID, recipient keys.ID) error {
	return w.writeSession(ctx, answer, sender, recipient, "answer")
}

func (w *Wormhole) readAnswer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient, "answer")
}

func (w *Wormhole) writeSession(ctx context.Context, addr *sctp.Addr, sender keys.ID, recipient keys.ID, id string) error {
	b, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	return w.hcl.PutEphemeral(ctx, sender, recipient, id, b)
}

func (w *Wormhole) readSession(ctx context.Context, sender keys.ID, recipient keys.ID, id string) (*sctp.Addr, error) {
	for {
		logger.Debugf("Get session (ephem/%s)...", id)
		b, err := w.hcl.GetEphemeral(ctx, sender, recipient, id)
		if err != nil {
			return nil, err
		}
		if b != nil {
			var addr sctp.Addr
			if err := json.Unmarshal(b, &addr); err != nil {
				return nil, err
			}
			return &addr, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			// Continue
		}
	}
}
