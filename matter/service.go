package matter

import (
	"context"
	"encoding/json"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
)

var _ MatterServer = &service{}

type service struct {
	UnimplementedMatterServer
	client *Client
	kr     Keyring
}

// Keyring for service.
type Keyring interface {
	EdX25519Key(kid keys.ID) (*keys.EdX25519Key, error)
}

// NewService is a service for Matter.
func NewService(client *Client, kr Keyring) MatterServer {
	return &service{
		client: client,
		kr:     kr,
	}
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	kid, err := keys.ParseID(req.KID)
	if err != nil {
		return nil, err
	}
	key, err := s.kr.EdX25519Key(kid)
	if err != nil {
		return nil, err
	}
	user, err := s.client.LoginWithKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		User: user,
	}, nil
}

func (s *service) CreateChannel(ctx context.Context, req *CreateChannelRequest) (*CreateChannelResponse, error) {
	channel, err := s.client.CreateChannel(ctx, req.Name, req.TeamID)
	if err != nil {
		return nil, err
	}
	return &CreateChannelResponse{
		Channel: channel,
	}, nil
}

func (s *service) TeamsForUser(ctx context.Context, req *TeamsForUserRequest) (*TeamsForUserResponse, error) {
	teams, err := s.client.TeamsForUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	return &TeamsForUserResponse{
		Teams: teams,
	}, nil
}

func (s *service) ChannelsForUser(ctx context.Context, req *ChannelsForUserRequest) (*ChannelsForUserResponse, error) {
	channels, err := s.client.ChannelsForUser(ctx, req.UserID, req.TeamID)
	if err != nil {
		return nil, err
	}
	return &ChannelsForUserResponse{
		Channels: channels,
	}, nil
}

func (s *service) Listen(server Matter_ListenServer) error {
	wsClient, err := s.client.NewWebSocketClient()
	if err != nil {
		return err
	}

	defer wsClient.Close()
	wsClient.Listen()

	go func() {
		for {
			logger.Debugf("Matter recv...")
			req, err := server.Recv()
			if err != nil {
				wsClient.Close()
				return
			}
			logger.Debugf("Matter recv req: %+v", req)
			// wsClient.SendMessage("", nil)
		}
	}()

	for {
		select {
		case event := <-wsClient.EventChannel:
			logger.Debugf(spew.Sdump(event))
			if event.Event == "posted" {
				postData, ok := event.Data["post"]
				if !ok {
					continue
				}
				var post Post
				if err := json.Unmarshal(postData.([]byte), &post); err != nil {
					log.Printf("Unrecognized post data\n")
					continue
				}

				if err := server.Send(&ListenEvent{
					Post: &post,
				}); err != nil {
					return err
				}
			}
		case resp := <-wsClient.ResponseChannel:
			logger.Debugf(spew.Sdump(resp))
		case _ = <-wsClient.PingTimeoutChannel:
			logger.Warningf("Matter websocket timed out")
			return nil
		default:
			return nil
		}
	}
}
