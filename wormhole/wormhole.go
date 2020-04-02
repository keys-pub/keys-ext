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

// TODO: If listening, after a little bit, retry the whole process from the
// start, in case the other side started over.

// ErrNoResponse error if offer not found for recipient.
var ErrNoResponse = errors.New("no response")

// ErrNoiseHandshakeTimeout if we timed out during handshake
var ErrNoiseHandshakeTimeout = errors.New("noise handshake timeout")

// ErrClosed if we recieve closed message.
var ErrClosed = errors.New("closed")

type Status string

const (
	SCTPHandshake  Status = "sctp-handshake"
	NoiseHandshake Status = "noise-handshake"
	Connected      Status = "open"
	Closed         Status = "closed"
)

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

	onStatus func(Status)
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
		rtc:      rtc,
		hcl:      hcl,
		ks:       ks,
		buf:      make([]byte, maxSize),
		onStatus: func(Status) {},
	}

	return w, nil
}

// Close wormhole.
func (w *Wormhole) Close() {
	w.Lock()
	defer w.Unlock()

	logger.Infof("Closing wormhole...")
	w.onStatus(Closed)
	w.onStatus = func(Status) {}

	if w.sender != "" {
		go func() {
			logger.Infof("Removing offer (if any)...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = w.hcl.DeleteMessage(ctx, w.sender, w.recipient, "wormhole", "offer")
		}()
		go func() {
			logger.Infof("Removing answer (if any)...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = w.hcl.DeleteMessage(ctx, w.sender, w.recipient, "wormhole", "answer")
		}()
	}

	go func() {
		logger.Infof("Sending close...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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

func (w *Wormhole) OnStatus(f func(Status)) {
	w.onStatus = f
}

func (w *Wormhole) Connect(ctx context.Context, sender keys.ID, recipient keys.ID, offer *sctp.Addr) error {
	logger.Infof("Wormhole connect...")

	w.sender = sender
	w.recipient = recipient

	ctxConnect, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	// While we're checking for an answer, continue to write the offer every 10
	// seconds, with a 15 second expire.
	go func() {
		for {
			if err := w.writeOffer(ctxConnect, offer, sender, recipient, false, time.Second*15); err != nil {
				return
			}
			select {
			case <-ctxConnect.Done():
				return
			case <-time.After(time.Second * 10):
				// Continue
			}
		}
	}()

	answer, err := w.readAnswer(ctxConnect, sender, recipient)
	if err != nil {
		return err
	}
	if answer == nil {
		return ErrNoResponse
	}

	cancel()

	w.onStatus(SCTPHandshake)

	if err := w.rtc.Connect(ctx, answer); err != nil {
		return err
	}

	return w.noiseHandshake(ctx, sender, recipient, true)
}

// FindInvite looks for an invite.
func (w *Wormhole) FindInvite(ctx context.Context, code string) (*api.InviteResponse, error) {
	// TODO: Brute force here is slow
	keys, err := w.ks.EdX25519Keys()
	if err != nil {
		return nil, err
	}
	var invite *api.InviteResponse
	for _, key := range keys {
		i, err := w.hcl.Invite(ctx, key.ID(), code)
		if err != nil {
			return nil, err
		}
		if i != nil {
			invite = i
			break
		}
	}
	if invite == nil {
		return nil, nil
	}
	return invite, nil
}

func (w *Wormhole) FindOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	sess, err := w.readOnce(ctx, sender, recipient, "offer")
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	return sess.Addr, nil
}

// CreateOffer creates an offer.
func (w *Wormhole) CreateOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.rtc.STUN(ctx, time.Second*10)
}

// CreateInvite creates an invite code for sender/recipient.
func (w *Wormhole) CreateInvite(ctx context.Context, sender keys.ID, recipient keys.ID) (string, error) {
	logger.Infof("Creating invite...")
	invite, err := w.hcl.CreateInvite(ctx, sender, recipient)
	if err != nil {
		return "", err
	}
	return invite.Code, nil
}

// CreateLocalOffer creates a local offer for testing.
func (w *Wormhole) CreateLocalOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.rtc.Local()
}

func (w *Wormhole) Listen(ctx context.Context, sender keys.ID, recipient keys.ID, offer *sctp.Addr) error {
	logger.Infof("Wormhole listen...")
	var answer *sctp.Addr
	if sctp.IsPrivateIP(offer.IP) {
		a, err := w.rtc.Local()
		if err != nil {
			return err
		}
		answer = a
	} else {
		a, err := w.rtc.STUN(ctx, time.Second*10)
		if err != nil {
			return err
		}
		answer = a
	}

	w.sender = sender
	w.recipient = recipient

	if err := w.writeAnswer(ctx, answer, sender, recipient); err != nil {
		return err
	}

	w.onStatus(SCTPHandshake)

	if err := w.rtc.ListenForPeer(ctx, offer); err != nil {
		return err
	}

	return w.noiseHandshake(ctx, sender, recipient, false)
}

func (w *Wormhole) noiseHandshake(ctx context.Context, sender keys.ID, recipient keys.ID, initiator bool) error {
	w.Lock()
	defer w.Unlock()

	w.onStatus(NoiseHandshake)

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

	noise, err := noise.NewNoise(senderKey.X25519Key(), recipientPublicKey.X25519PublicKey(), initiator)
	if err != nil {
		return err
	}

	noiseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// TODO: Test noise handshake timeout
	if initiator {
		out, err := noise.HandshakeWrite(nil)
		if err != nil {
			return err
		}
		if err := w.rtc.Write(noiseCtx, out); err != nil {
			return err
		}
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(noiseCtx, buf)
		if err != nil {
			return err
		}
		if _, err := noise.HandshakeRead(buf[:n]); err != nil {
			return err
		}
	} else {
		buf := make([]byte, 1024)
		n, err := w.rtc.Read(noiseCtx, buf)
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
		if err := w.rtc.Write(noiseCtx, out); err != nil {
			return err
		}
	}
	w.noise = noise

	logger.Infof("Wormhole connected.")
	w.onStatus(Connected)

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

func (w *Wormhole) writeOffer(ctx context.Context, offer *sctp.Addr, sender keys.ID, recipient keys.ID, genCode bool, expire time.Duration) error {
	return w.writeSession(ctx, offer, sender, recipient, "offer", genCode, expire)
}

func (w *Wormhole) readOffer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient, "offer")
}

func (w *Wormhole) writeAnswer(ctx context.Context, answer *sctp.Addr, sender keys.ID, recipient keys.ID) error {
	if err := w.writeSession(ctx, answer, sender, recipient, "answer", false, time.Hour); err != nil {
		return err
	}
	return nil
}

func (w *Wormhole) readAnswer(ctx context.Context, sender keys.ID, recipient keys.ID) (*sctp.Addr, error) {
	return w.readSession(ctx, sender, recipient, "answer")
}

type session struct {
	Addr *sctp.Addr `json:"addr"`
}

func (w *Wormhole) writeSession(ctx context.Context, addr *sctp.Addr, sender keys.ID, recipient keys.ID, typ string, genCode bool, expire time.Duration) error {
	sess := &session{
		Addr: addr,
	}
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	logger.Debugf("Writing session: %s (%s)", addr, typ)
	if err := w.hcl.SendMessage(ctx, sender, recipient, "wormhole", typ, b, expire); err != nil {
		return err
	}
	return nil
}

func (w *Wormhole) readSession(ctx context.Context, sender keys.ID, recipient keys.ID, typ string) (*sctp.Addr, error) {
	// TODO: Long polling?
	for {
		sess, err := w.readOnce(ctx, sender, recipient, typ)
		if err != nil {
			return nil, err
		}
		if sess != nil {
			return sess.Addr, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second * 3):
			// Continue
		}
	}
}

func (w *Wormhole) readOnce(ctx context.Context, sender keys.ID, recipient keys.ID, typ string) (*session, error) {
	logger.Debugf("Read session...")
	b, err := w.hcl.Message(ctx, sender, recipient, "wormhole", typ)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	if err := w.hcl.DeleteMessage(ctx, sender, recipient, "wormhole", typ); err != nil {
		return nil, err
	}
	var sess session
	if err := json.Unmarshal(b, &sess); err != nil {
		return nil, err
	}
	// logger.Debugf("Found addr: %s", addr)
	return &sess, nil
}
