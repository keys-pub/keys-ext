// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package client

import (
	"bytes"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Websocket constants.
const (
	SocketMaxMessageSize     = 8 * 1024 // 8KB
	PingTimeoutBufferSeconds = 5

	// WEBSOCKET_EVENT_TYPING                                   = "typing"
	// WEBSOCKET_EVENT_POSTED                                   = "posted"
	// WEBSOCKET_EVENT_POST_EDITED                              = "post_edited"
	// WEBSOCKET_EVENT_POST_DELETED                             = "post_deleted"
	// WEBSOCKET_EVENT_POST_UNREAD                              = "post_unread"
	// WEBSOCKET_EVENT_CHANNEL_CONVERTED                        = "channel_converted"
	// WEBSOCKET_EVENT_CHANNEL_CREATED                          = "channel_created"
	// WEBSOCKET_EVENT_CHANNEL_DELETED                          = "channel_deleted"
	// WEBSOCKET_EVENT_CHANNEL_RESTORED                         = "channel_restored"
	// WEBSOCKET_EVENT_CHANNEL_UPDATED                          = "channel_updated"
	// WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED                   = "channel_member_updated"
	// WEBSOCKET_EVENT_CHANNEL_SCHEME_UPDATED                   = "channel_scheme_updated"
	// WEBSOCKET_EVENT_DIRECT_ADDED                             = "direct_added"
	// WEBSOCKET_EVENT_GROUP_ADDED                              = "group_added"
	// WEBSOCKET_EVENT_NEW_USER                                 = "new_user"
	// WEBSOCKET_EVENT_ADDED_TO_TEAM                            = "added_to_team"
	// WEBSOCKET_EVENT_LEAVE_TEAM                               = "leave_team"
	// WEBSOCKET_EVENT_UPDATE_TEAM                              = "update_team"
	// WEBSOCKET_EVENT_DELETE_TEAM                              = "delete_team"
	// WEBSOCKET_EVENT_RESTORE_TEAM                             = "restore_team"
	// WEBSOCKET_EVENT_UPDATE_TEAM_SCHEME                       = "update_team_scheme"
	// WEBSOCKET_EVENT_USER_ADDED                               = "user_added"
	// WEBSOCKET_EVENT_USER_UPDATED                             = "user_updated"
	// WEBSOCKET_EVENT_USER_ROLE_UPDATED                        = "user_role_updated"
	// WEBSOCKET_EVENT_MEMBERROLE_UPDATED                       = "memberrole_updated"
	// WEBSOCKET_EVENT_USER_REMOVED                             = "user_removed"
	// WEBSOCKET_EVENT_PREFERENCE_CHANGED                       = "preference_changed"
	// WEBSOCKET_EVENT_PREFERENCES_CHANGED                      = "preferences_changed"
	// WEBSOCKET_EVENT_PREFERENCES_DELETED                      = "preferences_deleted"
	// WEBSOCKET_EVENT_EPHEMERAL_MESSAGE                        = "ephemeral_message"
	// WEBSOCKET_EVENT_STATUS_CHANGE                            = "status_change"
	// WEBSOCKET_EVENT_HELLO                                    = "hello"
	WebsocketAuthenticationChallenge = "authentication_challenge"
	// WEBSOCKET_EVENT_REACTION_ADDED                           = "reaction_added"
	// WEBSOCKET_EVENT_REACTION_REMOVED                         = "reaction_removed"
	// WEBSOCKET_EVENT_RESPONSE                                 = "response"
	// WEBSOCKET_EVENT_EMOJI_ADDED                              = "emoji_added"
	// WEBSOCKET_EVENT_CHANNEL_VIEWED                           = "channel_viewed"
	// WEBSOCKET_EVENT_PLUGIN_STATUSES_CHANGED                  = "plugin_statuses_changed"
	// WEBSOCKET_EVENT_PLUGIN_ENABLED                           = "plugin_enabled"
	// WEBSOCKET_EVENT_PLUGIN_DISABLED                          = "plugin_disabled"
	// WEBSOCKET_EVENT_ROLE_UPDATED                             = "role_updated"
	// WEBSOCKET_EVENT_LICENSE_CHANGED                          = "license_changed"
	// WEBSOCKET_EVENT_CONFIG_CHANGED                           = "config_changed"
	// WEBSOCKET_EVENT_OPEN_DIALOG                              = "open_dialog"
	// WEBSOCKET_EVENT_GUESTS_DEACTIVATED                       = "guests_deactivated"
	// WEBSOCKET_EVENT_USER_ACTIVATION_STATUS_CHANGE            = "user_activation_status_change"
	// WEBSOCKET_EVENT_RECEIVED_GROUP                           = "received_group"
	// WEBSOCKET_EVENT_RECEIVED_GROUP_ASSOCIATED_TO_TEAM        = "received_group_associated_to_team"
	// WEBSOCKET_EVENT_RECEIVED_GROUP_NOT_ASSOCIATED_TO_TEAM    = "received_group_not_associated_to_team"
	// WEBSOCKET_EVENT_RECEIVED_GROUP_ASSOCIATED_TO_CHANNEL     = "received_group_associated_to_channel"
	// WEBSOCKET_EVENT_RECEIVED_GROUP_NOT_ASSOCIATED_TO_CHANNEL = "received_group_not_associated_to_channel"
	// WEBSOCKET_EVENT_SIDEBAR_CATEGORY_CREATED                 = "sidebar_category_created"
	// WEBSOCKET_EVENT_SIDEBAR_CATEGORY_UPDATED                 = "sidebar_category_updated"
	// WEBSOCKET_EVENT_SIDEBAR_CATEGORY_DELETED                 = "sidebar_category_deleted"
	// WEBSOCKET_EVENT_SIDEBAR_CATEGORY_ORDER_UPDATED           = "sidebar_category_order_updated"
	// WEBSOCKET_WARN_METRIC_STATUS_RECEIVED                    = "warn_metric_status_received"
	// WEBSOCKET_WARN_METRIC_STATUS_REMOVED                     = "warn_metric_status_removed"
	// WEBSOCKET_EVENT_CLOUD_PAYMENT_STATUS_UPDATED             = "cloud_payment_status_updated"
)

type msgType int

const (
	msgTypeJSON msgType = iota + 1
	msgTypePong
)

type writeMessage struct {
	msgType msgType
	data    interface{}
}

const avgReadMsgSizeBytes = 1024

// WebSocketClient stores the necessary information required to
// communicate with a WebSocket endpoint.
// A client must read from PingTimeoutChannel, EventChannel and ResponseChannel to prevent
// deadlocks from occuring in the program.
type WebSocketClient struct {
	URL                string                  // The location of the server like "ws://localhost:8065"
	APIURL             string                  // The API location of the server like "ws://localhost:8065/api/v3"
	ConnectURL         string                  // The WebSocket URL to connect to like "ws://localhost:8065/api/v3/path/to/websocket"
	Conn               *websocket.Conn         // The WebSocket connection
	AuthToken          string                  // The token used to open the WebSocket connection
	Sequence           int64                   // The ever-incrementing sequence attached to each WebSocket action
	PingTimeoutChannel chan bool               // The channel used to signal ping timeouts
	EventChannel       chan *WebSocketEvent    // The channel used to receive various events pushed from the server. For example: typing, posted
	ResponseChannel    chan *WebSocketResponse // The channel used to receive responses for requests made to the server
	ListenError        error                   // A field that is set if there was an abnormal closure of the WebSocket connection
	writeChan          chan writeMessage

	pingTimeoutTimer *time.Timer
	quitPingWatchdog chan struct{}

	quitWriterChan chan struct{}
	resetTimerChan chan struct{}
	closed         int32
}

// NewWebSocketClient constructs a new WebSocket client with convenience
// methods for talking to the server.
func NewWebSocketClient(url, authToken string) (*WebSocketClient, error) {
	return NewWebSocketClientWithDialer(websocket.DefaultDialer, url, authToken)
}

// WebSocketEvent ...
type WebSocketEvent struct {
	Event           string
	Data            map[string]interface{}
	Broadcast       *WebsocketBroadcast
	Sequence        int64
	precomputedJSON *precomputedWebSocketEventJSON
}

type precomputedWebSocketEventJSON struct {
	Event     json.RawMessage
	Data      json.RawMessage
	Broadcast json.RawMessage
}

// WebsocketBroadcast ...
type WebsocketBroadcast struct {
	OmitUsers map[string]bool `json:"omit_users"` // broadcast is omitted for users listed here
	UserID    string          `json:"user_id"`    // broadcast only occurs for this user
	ChannelID string          `json:"channel_id"` // broadcast only occurs for users in this channel
	TeamID    string          `json:"team_id"`    // broadcast only occurs for users in this team
}

// WebSocketResponse represents a response received through the WebSocket
// for a request made to the server. This is available through the ResponseChannel
// channel in WebSocketClient.
type WebSocketResponse struct {
	Status   string                 `json:"status"`              // The status of the response. For example: OK, FAIL.
	SeqReply int64                  `json:"seq_reply,omitempty"` // A counter which is incremented for every response sent.
	Data     map[string]interface{} `json:"data,omitempty"`      // The data contained in the response.
	Error    error                  `json:"error,omitempty"`     // A field that is set if any error has occurred.
}

// NewWebSocketClientWithDialer constructs a new WebSocket client with convenience
// methods for talking to the server using a custom dialer.
func NewWebSocketClientWithDialer(dialer *websocket.Dialer, url, authToken string) (*WebSocketClient, error) {
	apiURL := url + "/api/v4"
	connectURL := apiURL + "/websocket"

	conn, _, err := dialer.Dial(connectURL, nil)
	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		URL:                url,
		APIURL:             apiURL,
		ConnectURL:         connectURL,
		Conn:               conn,
		AuthToken:          authToken,
		Sequence:           1,
		PingTimeoutChannel: make(chan bool, 1),
		EventChannel:       make(chan *WebSocketEvent, 100),
		ResponseChannel:    make(chan *WebSocketResponse, 100),
		writeChan:          make(chan writeMessage),
		quitPingWatchdog:   make(chan struct{}),
		quitWriterChan:     make(chan struct{}),
		resetTimerChan:     make(chan struct{}),
	}

	client.configurePingHandling()
	go client.writer()

	client.SendMessage(WebsocketAuthenticationChallenge, map[string]interface{}{"token": authToken})

	return client, nil
}

// Connect creates a websocket connection with the given ConnectUrl.
// This is racy and error-prone should not be used. Use any of the New* functions to create a websocket.
func (wsc *WebSocketClient) Connect() error {
	return wsc.ConnectWithDialer(websocket.DefaultDialer)
}

// ConnectWithDialer creates a websocket connection with the given ConnectUrl using the dialer.
// This is racy and error-prone and should not be used. Use any of the New* functions to create a websocket.
func (wsc *WebSocketClient) ConnectWithDialer(dialer *websocket.Dialer) error {
	var err error
	wsc.Conn, _, err = dialer.Dial(wsc.ConnectURL, nil)
	if err != nil {
		return err
	}
	// Super racy and should not be done anyways.
	// All of this needs to be redesigned for v6.
	wsc.configurePingHandling()
	// If it has been closed before, we just restart the writer.
	if atomic.CompareAndSwapInt32(&wsc.closed, 1, 0) {
		wsc.writeChan = make(chan writeMessage)
		wsc.quitWriterChan = make(chan struct{})
		go wsc.writer()
		wsc.resetTimerChan = make(chan struct{})
		wsc.quitPingWatchdog = make(chan struct{})
	}

	wsc.EventChannel = make(chan *WebSocketEvent, 100)
	wsc.ResponseChannel = make(chan *WebSocketResponse, 100)

	wsc.SendMessage(WebsocketAuthenticationChallenge, map[string]interface{}{"token": wsc.AuthToken})

	return nil
}

// Close closes the websocket client. It is recommended that a closed client should not be
// reused again. Rather a new client should be created anew.
func (wsc *WebSocketClient) Close() {
	// CAS to 1 and proceed. Return if already 1.
	if !atomic.CompareAndSwapInt32(&wsc.closed, 0, 1) {
		return
	}
	wsc.quitWriterChan <- struct{}{}
	close(wsc.writeChan)
	// We close the connection, which breaks the reader loop.
	// Then we let the defer block in the reader do further cleanup.
	wsc.Conn.Close()
}

// TODO: un-export the Conn so that Write methods go through the writer
func (wsc *WebSocketClient) writer() {
	for {
		select {
		case msg := <-wsc.writeChan:
			switch msg.msgType {
			case msgTypeJSON:
				wsc.Conn.WriteJSON(msg.data)
			case msgTypePong:
				wsc.Conn.WriteMessage(websocket.PongMessage, []byte{})
			}
		case <-wsc.quitWriterChan:
			return
		}
	}
}

// Listen starts the read loop of the websocket client.
func (wsc *WebSocketClient) Listen() {
	// This loop can exit in 2 conditions:
	// 1. Either the connection breaks naturally.
	// 2. Close was explicitly called, which closes the connection manually.
	//
	// Due to the way the API is written, there is a requirement that a client may NOT
	// call Listen at all and can still call Close and Connect.
	// Therefore, we let the cleanup of the reader stuff rely on closing the connection
	// and then we do the cleanup in the defer block.
	//
	// First, we close some channels and then CAS to 1 and proceed to close the writer chan also.
	// This is needed because then the defer clause does not double-close the writer when (2) happens.
	// But if (1) happens, we set the closed bit, and close the rest of the stuff.
	go func() {
		defer func() {
			close(wsc.EventChannel)
			close(wsc.ResponseChannel)
			close(wsc.quitPingWatchdog)
			close(wsc.resetTimerChan)
			// We CAS to 1 and proceed.
			if !atomic.CompareAndSwapInt32(&wsc.closed, 0, 1) {
				return
			}
			wsc.quitWriterChan <- struct{}{}
			close(wsc.writeChan)
			wsc.Conn.Close() // This can most likely be removed. Needs to be checked.
		}()

		var buf bytes.Buffer
		buf.Grow(avgReadMsgSizeBytes)

		for {
			// Reset buffer.
			buf.Reset()
			_, r, err := wsc.Conn.NextReader()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
					wsc.ListenError = err
				}
				return
			}
			// Use pre-allocated buffer.
			_, err = buf.ReadFrom(r)
			if err != nil {
				// This should use a different error ID, but en.json is not imported anyways.
				// It's a different bug altogether but we let it be for now.
				// See MM-24520.
				wsc.ListenError = err
				return
			}

			var event WebSocketEvent
			if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
				continue
			}

			if event.Event != "" {
				wsc.EventChannel <- &event
				continue
			}

			var response WebSocketResponse
			if err := json.Unmarshal(buf.Bytes(), &response); err == nil && response.Status != "" {
				wsc.ResponseChannel <- &response
				continue
			}
		}
	}()
}

// WebSocketRequest represents a request made to the server through a websocket.
type WebSocketRequest struct {
	// Client-provided fields
	Seq    int64                  `json:"seq"`    // A counter which is incremented for every request made.
	Action string                 `json:"action"` // The action to perform for a request. For example: get_statuses, user_typing.
	Data   map[string]interface{} `json:"data"`   // The metadata for an action.
}

// SendMessage sens a message.
func (wsc *WebSocketClient) SendMessage(action string, data map[string]interface{}) {
	req := &WebSocketRequest{}
	req.Seq = wsc.Sequence
	req.Action = action
	req.Data = data

	wsc.Sequence++
	wsc.writeChan <- writeMessage{
		msgType: msgTypeJSON,
		data:    req,
	}
}

// UserTyping will push a user_typing event out to all connected users
// who are in the specified channel
func (wsc *WebSocketClient) UserTyping(channelID, parentID string) {
	data := map[string]interface{}{
		"channel_id": channelID,
		"parent_id":  parentID,
	}

	wsc.SendMessage("user_typing", data)
}

// GetStatuses will return a map of string statuses using user id as the key
func (wsc *WebSocketClient) GetStatuses() {
	wsc.SendMessage("get_statuses", nil)
}

// GetStatusesByIds will fetch certain user statuses based on ids and return
// a map of string statuses using user id as the key
func (wsc *WebSocketClient) GetStatusesByIds(userIds []string) {
	data := map[string]interface{}{
		"user_ids": userIds,
	}
	wsc.SendMessage("get_statuses_by_ids", data)
}

func (wsc *WebSocketClient) configurePingHandling() {
	wsc.Conn.SetPingHandler(wsc.pingHandler)
	wsc.pingTimeoutTimer = time.NewTimer(time.Second * (60 + PingTimeoutBufferSeconds))
	go wsc.pingWatchdog()
}

func (wsc *WebSocketClient) pingHandler(appData string) error {
	if atomic.LoadInt32(&wsc.closed) == 1 {
		return nil
	}
	wsc.resetTimerChan <- struct{}{}
	wsc.writeChan <- writeMessage{
		msgType: msgTypePong,
	}
	return nil
}

// pingWatchdog is used to send values to the PingTimeoutChannel whenever a timeout occurs.
// We use the resetTimerChan from the pingHandler to pass the signal, and then reset the timer
// after draining it. And if the timer naturally expires, we also extend it to prevent it from
// being deadlocked when the resetTimerChan case runs. Because timer.Stop would return false,
// and the code would be forever stuck trying to read from C.
func (wsc *WebSocketClient) pingWatchdog() {
	for {
		select {
		case <-wsc.resetTimerChan:
			if !wsc.pingTimeoutTimer.Stop() {
				<-wsc.pingTimeoutTimer.C
			}
			wsc.pingTimeoutTimer.Reset(time.Second * (60 + PingTimeoutBufferSeconds))

		case <-wsc.pingTimeoutTimer.C:
			wsc.PingTimeoutChannel <- true
			wsc.pingTimeoutTimer.Reset(time.Second * (60 + PingTimeoutBufferSeconds))
		case <-wsc.quitPingWatchdog:
			return
		}
	}
}
