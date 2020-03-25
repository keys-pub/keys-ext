package wormhole

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/noise"
	"github.com/keys-pub/keysd/http/api"
	httpclient "github.com/keys-pub/keysd/http/client"
	"github.com/keys-pub/keysd/wormhole/sctp"
	"github.com/pkg/errors"
)

// ErrNoResponse error if offer not found for recipient.
var ErrNoResponse = errors.New("no response")

// ErrNoiseHandshakeTimeout if we timed out during handshake
var ErrNoiseHandshakeTimeout = errors.New("noise handshake timeout")

// ErrClosed if we recieve closed message.
var ErrClosed = errors.New("closed")

// ErrInviteNotFound if invite code not found
var ErrInviteNotFound = errors.New("invite code not found")

// Wormhole for connecting two participants using webrtc, noise and
// keys.pub.
type Wormhole struct {
	sync.Mutex
	rtc   *sctp.Client
	hcl   *httpclient.Client
	ks    *keys.Keystore
	noise *noise.Noise

	sender    keys.ID
	recipient keys.ID

	buf []byte

	onConnect func()
	onClose   func()
}

// maxSize of write/read.
// The max content size may be less than this because of header bytes.
const maxSize = 16 * 1024

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
		ks:        ks,
		buf:       make([]byte, maxSize),
		onConnect: func() {},
		onClose:   func() {},
	}

	return w, nil
}

// Close wormhole.
func (w *Wormhole) Close() {
	w.Lock()
	defer w.Unlock()

	w.onClose()
	w.onClose = func() {}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = w.writeClosed(ctx)

		time.Sleep(time.Second)
		w.rtc.Close()
	}()
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

func (w *Wormhole) Connect(ctx context.Context, sender keys.ID, recipient keys.ID, offer *sctp.Addr) error {
	logger.Infof("Connect")
	if err := w.connect(ctx, sender, recipient, offer); err != nil {
		return err
	}
	return w.start(ctx, sender, recipient, true)
}

func (w *Wormhole) ListenByInvite(ctx context.Context, code string) error {
	// TODO: Brute force here is slow
	keys, err := w.ks.EdX25519Keys()
	if err != nil {
		return err
	}
	var invite *api.InviteResponse
	for _, key := range keys {
		i, err := w.hcl.GetInvite(ctx, key.ID(), code)
		if err != nil {
			return err
		}
		if i != nil {
			invite = i
			break
		}
	}
	if invite == nil {
		return ErrInviteNotFound
	}
	sender, recipient := invite.Recipient, invite.Sender

	offer, err := w.FindOffer(ctx, sender, recipient)
	if err != nil {
		return err
	}

	return w.Listen(ctx, sender, recipient, offer)
}

func (w *Wormhole) FindOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readOnce(ctx, sender, recipient)
}

func (w *Wormhole) CreateOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, string, error) {
	offer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return nil, "", err
	}
	code, err := w.writeOffer(ctx, offer, sender, recipient)
	if err != nil {
		return nil, "", err
	}
	return offer, code, nil
}

func (w *Wormhole) connect(ctx context.Context, sender keys.ID, recipient keys.ID, offer *sctp.Addr) error {
	logger.Infof("Wormhole connect...")

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

func (w *Wormhole) Listen(ctx context.Context, sender keys.ID, recipient keys.ID, offer *sctp.Addr) error {
	logger.Infof("Wormhole listen...")
	answer, err := w.rtc.STUN(ctx, time.Second*10)
	if err != nil {
		return err
	}
	if err := w.writeAnswer(ctx, answer, sender, recipient); err != nil {
		return err
	}

	if err := w.rtc.Listen(ctx, offer); err != nil {
		return err
	}

	return w.start(ctx, sender, recipient, false)
}

func (w *Wormhole) start(ctx context.Context, sender keys.ID, recipient keys.ID, initiator bool) error {
	w.Lock()
	defer w.Unlock()

	if w.noise != nil {
		return errors.Errorf("wormhole already started")
	}

	senderKey, err := w.ks.EdX25519Key(sender)
	if err != nil {
		return err
	}
	if senderKey == nil {
		return keys.NewErrNotFound(sender.String())
	}
	recipientPublicKey, err := w.ks.EdX25519PublicKey(recipient)
	if err != nil {
		return err
	}
	if recipientPublicKey == nil {
		return keys.NewErrNotFound(recipientPublicKey.String())
	}

	w.sender = sender
	w.recipient = recipient

	noise, err := noise.NewNoise(senderKey.X25519Key(), recipientPublicKey.X25519PublicKey(), initiator)
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

// Write data.
func (w *Wormhole) Write(ctx context.Context, b []byte) error {
	if w.noise == nil {
		return errors.Errorf("no channel (noise)")
	}
	if len(b) > maxSize {
		return errors.Errorf("write exceeds max size")
	}
	encrypted, err := w.noise.Encrypt(nil, nil, b)
	if err != nil {
		return err
	}
	return w.rtc.Write(ctx, encrypted)
}

// Read.
func (w *Wormhole) Read(ctx context.Context) ([]byte, error) {
	n, err := w.rtc.Read(ctx, w.buf)
	if err != nil {
		return nil, err
	}

	logger.Infof("Wormhole read (%d)", n)
	decrypted, err := w.noise.Decrypt(nil, nil, w.buf[:n])
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// NewID creates new ID for wormhole messages.
func NewID() string {
	return encoding.MustEncode(keys.Rand32()[:], encoding.Base62)
}

// WriteMessage writes a message.
func (w *Wormhole) WriteMessage(ctx context.Context, id string, b []byte, contentType ContentType) (*Message, error) {
	if len(b) > maxSize-33 {
		return nil, errors.Errorf("write exceeds max size")
	}

	decid, err := encoding.Decode(id, encoding.Base62)
	if err != nil {
		return nil, err
	}
	if len(decid) != 32 {
		return nil, errors.Errorf("invalid id for wormhole write, 32 != %d", len(decid))
	}

	out := append([]byte{msgByte}, decid[:]...)
	out = append(out, b...)
	if err := w.Write(ctx, out); err != nil {
		return nil, err
	}

	msg := &Message{
		ID:        id,
		Sender:    w.sender,
		Recipient: w.recipient,
		Content: &Content{
			Data: b,
			Type: contentType,
		},
		Type: Pending,
	}

	return msg, nil
}

const msgByte byte = 0x01
const ackByte byte = 0x02
const closedByte byte = 0xDE

// ReadMessage reads a message.
// If ack, will send an ack (unless this message is an ack).
// If we received a message that the recipient closed, we return ErrClosed.
func (w *Wormhole) ReadMessage(ctx context.Context, ack bool) (*Message, error) {
	b, err := w.Read(ctx)
	if err != nil {
		return nil, err
	}

	typ := b[0]
	msgType := Sent
	switch typ {
	case ackByte:
		msgType = Ack
	case closedByte:
		w.Close()
		return nil, ErrClosed
	}

	if len(b) < 33 {
		return nil, errors.Errorf("not enough bytes for a message")
	}

	bid := keys.Bytes32(b[1:33])
	payload := b[33:]

	contentType := BinaryContent
	if utf8.Valid(payload) {
		contentType = UTF8Content
	}

	id := encoding.MustEncode(bid[:], encoding.Base62)

	msg := &Message{
		ID:        id,
		Sender:    w.recipient,
		Recipient: w.sender,
		Content: &Content{
			Data: payload,
			Type: contentType,
		},
		Type: msgType,
	}

	if msgType != Ack {
		if err := w.writeAck(ctx, bid); err != nil {
			return nil, errors.Wrapf(err, "failed to ack")
		}
	}

	return msg, nil
}

func (w *Wormhole) writeAck(ctx context.Context, id *[32]byte) error {
	return w.Write(ctx, append([]byte{ackByte}, id[:]...))
}

func (w *Wormhole) writeClosed(ctx context.Context) error {
	return w.Write(ctx, []byte{closedByte, 0xAD, 0xBE, 0xEF})
}

func (w *Wormhole) writeOffer(ctx context.Context, offer *sctp.Addr, sender keys.ID, recipient keys.ID) (string, error) {
	return w.writeSession(ctx, offer, sender, recipient, true)
}

func (w *Wormhole) readOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient)
}

func (w *Wormhole) writeAnswer(ctx context.Context, answer *sctp.Addr, sender keys.ID, recipient keys.ID) error {
	_, err := w.writeSession(ctx, answer, sender, recipient, false)
	return err
}

func (w *Wormhole) readAnswer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient)
}

func (w *Wormhole) writeSession(ctx context.Context, addr *sctp.Addr, sender keys.ID, recipient keys.ID, genCode bool) (string, error) {
	b, err := json.Marshal(addr)
	if err != nil {
		return "", err
	}
	// logger.Debugf("Put ephem: %s", addr)
	resp, err := w.hcl.PutEphemeral(ctx, sender, recipient, b, genCode)
	if err != nil {
		return "", err
	}
	return resp.Code, nil
}

func (w *Wormhole) readSession(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	for {
		addr, err := w.readOnce(ctx, sender, recipient)
		if err != nil {
			return nil, err
		}
		if addr != nil {
			return addr, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second * 2):
			// Continue
		}
	}
}

func (w *Wormhole) readOnce(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	logger.Debugf("Read ephem...")
	b, err := w.hcl.GetEphemeral(ctx, sender, recipient)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	var addr sctp.Addr
	if err := json.Unmarshal(b, &addr); err != nil {
		return nil, err
	}
	// logger.Debugf("Found addr: %s", addr)
	return &addr, nil
}
