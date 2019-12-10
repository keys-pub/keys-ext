package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// Pull (RPC)
func (s *service) Pull(ctx context.Context, req *PullRequest) (*PullResponse, error) {
	if req.All {
		if req.KID != "" || req.User != "" {
			return nil, errors.Errorf("all specified with other arguments")
		}

		kids, err := s.pullStatements(ctx)
		if err != nil {
			return nil, err
		}
		return &PullResponse{KIDs: kids}, nil
	}

	if req.KID != "" {
		kid, err := keys.ParseID(req.KID)
		if err != nil {
			return nil, err
		}
		ok, err := s.pull(ctx, kid)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.Errorf("%s not found", req.KID)
		}
		return &PullResponse{KIDs: []string{kid.String()}}, nil
	} else if req.User != "" {
		usr, err := s.findUserByName(ctx, req.User)
		if err != nil {
			return nil, err
		}
		if usr == nil {
			return &PullResponse{}, nil
		}
		ok, err := s.pull(ctx, usr.KID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.Errorf("%s not found", req.User)
		}
		return &PullResponse{KIDs: []string{usr.KID.String()}}, nil
	}

	// Update existing if no kid or user specified
	pulled := []string{}
	kids, err := s.kidsSet(true)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load kids")
	}
	for _, kid := range kids.IDs() {
		ok, err := s.pull(ctx, kid)
		if err != nil {
			return nil, err
		}
		if !ok {
			// TODO: Report missing
			continue
		}
		pulled = append(pulled, kid.String())
	}
	return &PullResponse{KIDs: pulled}, nil
}

func (s *service) pull(ctx context.Context, kid keys.ID) (bool, error) {
	if s.remote == nil {
		return false, errors.Errorf("no remote set")
	}
	logger.Infof("Pull sigchain %s", kid)
	resp, err := s.remote.Sigchain(kid)
	if err != nil {
		return false, err
	}
	if resp == nil {
		logger.Infof("No sigchain for %s", kid)
		return false, nil
	}
	logger.Infof("Received sigchain %s, len=%d", kid, len(resp.Statements))
	for _, st := range resp.Statements {
		if err := s.db.Set(ctx, st.KeyPath(), st.Bytes()); err != nil {
			return false, err
		}
		if err := s.saveResource(ctx, st.KeyPath(), resp.MetadataFor(st)); err != nil {
			return false, err
		}

	}
	return true, nil
}

func (s *service) pullStatements(ctx context.Context) ([]string, error) {
	logger.Infof("Pull statements...")
	versionPath := keys.Path("versions", "sigchains")
	e, err := s.db.Get(ctx, versionPath)
	if err != nil {
		return nil, err
	}

	version := ""
	if e != nil {
		version = string(e.Data)
	}
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	resp, err := s.remote.Sigchains(version)
	if err != nil {
		return nil, err
	}
	kids := keys.NewStringSet()
	for _, st := range resp.Statements {
		if err := s.db.Set(ctx, st.KeyPath(), st.Bytes()); err != nil {
			return nil, err
		}
		if err := s.saveResource(ctx, st.KeyPath(), resp.MetadataFor(st)); err != nil {
			return nil, err
		}
		kids.Add(st.KID.String())
	}
	if err := s.db.Set(ctx, versionPath, []byte(resp.Version)); err != nil {
		return nil, err
	}
	return kids.Strings(), nil
}

func (s *service) pullMessages(ctx context.Context, kid keys.ID) error {
	key, err := s.ks.Key(kid)
	if err != nil {
		return err
	}
	if key == nil {
		return keys.NewErrNotFound(kid, keys.KeyType)
	}
	logger.Infof("Pull messages...")
	versionPath := keys.Path("versions", fmt.Sprintf("messages-%s", kid))
	e, err := s.db.Get(ctx, versionPath)
	if err != nil {
		return err
	}
	version := ""
	if e != nil {
		version = string(e.Data)
	}
	if s.remote == nil {
		return errors.Errorf("no remote set")
	}
	resp, err := s.remote.Messages(key, version)
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	logger.Infof("Received %d messages", len(resp.Messages))
	for _, msg := range resp.Messages {
		md := resp.MetadataFor(msg)
		ts := 9223372036854775807 - keys.TimeToMillis(md.CreatedAt)
		pathKey := fmt.Sprintf("messages-%s", key.ID())
		pathVal := fmt.Sprintf("%d-%s", ts, msg.ID)
		path := keys.Path(pathKey, pathVal)
		if err := s.db.Set(ctx, path, msg.Data); err != nil {
			return err
		}
		if err := s.saveResource(ctx, path, resp.MetadataFor(msg)); err != nil {
			return err
		}
	}
	if err := s.db.Set(ctx, versionPath, []byte(resp.Version)); err != nil {
		return err
	}
	return nil
}

func (s *service) vaultExists(key keys.Key) (bool, error) {
	resp, err := s.remote.Vault(key, "")
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}
	return true, nil
}

// func (s *service) pullVault(key keys.Key, full bool) (bool, error) {
// 	logger.Infof("Pull vault...")
// 	version := ""
// 	versionPath := keys.Path("versions", fmt.Sprintf("vault-%s", key.ID()))
// 	if !full {
// 		e, err := s.db.Get(context.TODO(), versionPath)
// 		if err != nil {
// 			return false, err
// 		}
// 		if e != nil {
// 			version = string(e.Data)
// 		}
// 	}
// 	if s.remote == nil {
// 		return false, errors.Errorf("no remote set")
// 	}
// 	resp, err := s.remote.Vault(key, version)
// 	if err != nil {
// 		return false, err
// 	}
// 	if resp == nil {
// 		return false, nil
// 	}
// 	logger.Infof("Received %d", len(resp.Items))
// 	for _, item := range resp.Items {
// 		// if err := s.db.Put(ctx, item.Path, item.Data, &db.Metadata{CreatedAt: item.CreatedAt}); err != nil {
// 		// 	return err
// 		// }
// 		if _, err := s.pullSigchain(item.ID); err != nil {
// 			return false, err
// 		}
// 	}
// 	if err := s.db.Put(context.TODO(), versionPath, []byte(resp.Version), &db.Metadata{CreatedAt: time.Now()}); err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }

type resource struct {
	Path     string       `json:"path"`
	SavedAt  time.Time    `json:"savedAt"`
	Metadata api.Metadata `json:"md"`
}

func (s *service) saveResource(ctx context.Context, path string, md api.Metadata) error {
	resource := &resource{
		Path:     path,
		SavedAt:  s.Now(),
		Metadata: md,
	}
	b, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	rp := keys.Path(".resource", path)
	logger.Debugf("Saving pull resource: %s", rp)
	if err := s.db.Set(ctx, rp, b); err != nil {
		return err
	}
	return nil
}

func (s *service) loadResource(ctx context.Context, path string) (*resource, error) {
	rp := keys.Path(".resource", path)
	logger.Debugf("Load pull resource: %s", rp)
	doc, err := s.db.Get(ctx, rp)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var res resource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
