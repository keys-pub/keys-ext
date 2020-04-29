package service

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

// Preferences (RPC) ...
func (s *service) Preferences(ctx context.Context, req *PreferencesRequest) (*PreferencesResponse, error) {
	prefs, err := s.prefs(ctx)
	if err != nil {
		return nil, err
	}

	return &PreferencesResponse{
		Prefs: prefs,
	}, nil
}

func (s *service) prefs(ctx context.Context) ([]*Pref, error) {
	if s.db == nil {
		return nil, errors.Errorf("db not available")
	}

	doc, err := s.db.Get(ctx, "prefs")
	if err != nil {
		return nil, err
	}

	var prefs []*Pref
	if doc != nil {
		if err := json.Unmarshal(doc.Data, &prefs); err != nil {
			return nil, errors.Errorf("failed to get prefs from db")
		}
	}

	// Ensure empty
	if len(prefs) == 0 {
		prefs = []*Pref{}
	}

	return prefs, nil
}

func (s *service) savePrefs(ctx context.Context, prefs []*Pref) error {
	if s.db == nil {
		return errors.Errorf("db not available")
	}

	b, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	if err := s.db.Set(ctx, "prefs", b); err != nil {
		return errors.Wrapf(err, "failed to save prefs")
	}
	return nil
}

// PreferenceSet (RPC) ...
func (s *service) PreferenceSet(ctx context.Context, req *PreferenceSetRequest) (*PreferenceSetResponse, error) {
	prefs, err := s.prefs(ctx)
	if err != nil {
		return nil, err
	}

	updated := false
	for _, pref := range prefs {
		if pref.Key == req.Pref.Key {
			pref.Value = req.Pref.Value
			updated = true
		}
	}

	if !updated {
		prefs = append(prefs, req.Pref)
	}

	if err := s.savePrefs(ctx, prefs); err != nil {
		return nil, err
	}

	return &PreferenceSetResponse{}, nil
}
