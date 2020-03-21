package wormhole

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
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

	onOpen    func()
	onClose   func()
	onMessage func(b []byte)
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
		onOpen:    func() {},
		onClose:   func() {},
		onMessage: func(b []byte) {},
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

func (w *Wormhole) OnOpen(f func()) {
	w.onOpen = f
}

func (w *Wormhole) OnClose(f func()) {
	w.onClose = f
}

func (w *Wormhole) OnMessage(f func(b []byte)) {
	w.onMessage = f
}

func (w *Wormhole) messageLn(b []byte) {
	logger.Infof("Message (%d)", len(b))
	decrypted, err := w.noise.Decrypt(nil, nil, b)
	if err != nil {
		logger.Errorf("Failed to decrypt message: %s", err)
		return
	}
	w.onMessage(decrypted)
}

func (w *Wormhole) openLn() {
	w.onOpen()
}

func (w *Wormhole) Start(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey) error {
	w.Lock()
	defer w.Unlock()

	if w.noise != nil {
		return errors.Errorf("wormhole already started")
	}
	// Initiator is whichever ID is less than
	initiator := sender.ID() < recipient.ID()
	logger.Infof("Initator: %t", initiator)

	if initiator {
		if err := w.connect(ctx, sender, recipient); err != nil {
			return err
		}
	} else {
		if err := w.listen(ctx, sender, recipient); err != nil {
			return err
		}
	}

	noise, err := noise.NewNoise(sender.X25519Key(), recipient.X25519PublicKey(), initiator)
	if err != nil {
		return err
	}
	w.noise = noise

	if initiator {
		out, err := w.noise.HandshakeWrite(nil)
		if err != nil {
			return err
		}
		if err := w.rtc.Write(out); err != nil {
			return err
		}
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(buf)
		if err != nil {
			return err
		}
		if _, err := w.noise.HandshakeRead(buf[:n]); err != nil {
			return err
		}
	} else {
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(buf)
		if err != nil {
			return err
		}
		if _, err := w.noise.HandshakeRead(buf[:n]); err != nil {
			return err
		}
		out, err := w.noise.HandshakeWrite(nil)
		if err != nil {
			return err
		}
		if err := w.rtc.Write(out); err != nil {
			return err
		}
	}

	logger.Infof("Started")
	w.openLn()

	// Read
	go func() {
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(buf)
		if err != nil {
			logger.Errorf("Read error: %v", err)
		}
		w.messageLn(buf[:n])
	}()

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
	return w.rtc.Write(encrypted)
}

func (w *Wormhole) connect(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey) error {
	logger.Infof("Connect...")
	offer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return err
	}
	if err := writeOffer(offer); err != nil {
		return err
	}

	answer, err := readAnswer()
	if err != nil {
		return err
	}
	if answer == nil {
		return ErrNoResponse
	}

	if err := w.rtc.Handshake(ctx, answer, time.Second*5); err != nil {
		return err
	}

	if err := w.rtc.Connect(answer); err != nil {
		return err
	}

	return nil
}

func (w *Wormhole) listen(ctx context.Context, sender *keys.EdX25519Key, recipient *keys.EdX25519PublicKey) error {
	logger.Infof("Listen...")
	answer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return err
	}
	if err := writeAnswer(answer); err != nil {
		return err
	}

	offer, err := readOffer()
	if err != nil {
		return err
	}
	if offer == nil {
		return ErrNoResponse
	}

	if err := w.rtc.Handshake(ctx, offer, time.Second*5); err != nil {
		return err
	}

	if err := w.rtc.Listen(context.TODO(), offer); err != nil {
		log.Fatal(err)
	}

	return nil
}

func writeOffer(offer *sctp.Addr) error {
	return sendAddr(offer, "https://keys.pub/relay/offer")
}

func readOffer() (*sctp.Addr, error) {
	return readAddr("https://keys.pub/relay/offer")
}

func writeAnswer(answer *sctp.Addr) error {
	return sendAddr(answer, "https://keys.pub/relay/answer")
}

func readAnswer() (*sctp.Addr, error) {
	return readAddr("https://keys.pub/relay/answer")
}

func sendAddr(addr *sctp.Addr, url string) error {
	b, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	logger.Infof("Send %s: %s", url, string(b))
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func readAddr(url string) (*sctp.Addr, error) {
	for {
		logger.Infof("Look %s...", url)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var addr sctp.Addr
			if err = json.NewDecoder(resp.Body).Decode(&addr); err != nil {
				return nil, err
			}
			logger.Infof("Got address.")
			return &addr, nil
		} else if resp.StatusCode == 404 {
			time.Sleep(time.Second)
		} else {
			return nil, errors.Errorf("Failed to get offer %d", resp.StatusCode)
		}
	}
}
