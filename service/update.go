package service

import (
	"context"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/pkg/errors"
)

func (s *service) checkKeys(ctx context.Context) error {
	logger.Infof("Checking keys...")
	pks, err := s.vault.EdX25519PublicKeys()
	if err != nil {
		return errors.Wrapf(err, "failed to list public keys")
	}
	for _, pk := range pks {
		if err := ctx.Err(); err != nil {
			return err
		}
		// We only need to do this on key creation or after a sigchain update,
		// but old versions have never sigchain indexed before, so we'll do this
		// here every time.
		if err := s.scs.Index(pk.ID()); err != nil {
			return err
		}

		if err := s.checkForExpiredKey(ctx, pk.ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) checkForExpiredKey(ctx context.Context, kid keys.ID) error {
	res, err := s.users.Get(ctx, kid)
	if err != nil {
		return err
	}
	// Check if expired, and then update.
	// If we don't have a local result, we don't update.
	for _, r := range res {
		// If not OK, check every "userCheckFailureExpire", otherwise check every "userCheckExpire".
		now := s.clock.Now()
		if (r.Status != user.StatusOK && r.IsTimestampExpired(now, userCheckFailureExpire)) ||
			r.IsTimestampExpired(now, userCheckExpire) {
			if _, _, err := s.update(ctx, kid); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) updateAllKeys(ctx context.Context) error {
	logger.Infof("Updating keys...")
	pks, err := s.vault.EdX25519PublicKeys()
	if err != nil {
		return err
	}
	for _, pk := range pks {
		if _, _, err := s.update(ctx, pk.ID()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) update(ctx context.Context, kid keys.ID) (bool, []*user.Result, error) {
	logger.Infof("Update %s", kid)

	resp, err := s.client.Sigchain(ctx, kid)
	if err != nil {
		return false, nil, err
	}
	if resp != nil {
		// TODO: Check that our existing statements haven't changed or disappeared
		logger.Infof("Received sigchain %s, len=%d", kid, len(resp.Statements))

		sc := keys.NewSigchain(kid)
		if err := sc.AddAll(resp.Statements); err != nil {
			return false, nil, err
		}
		if err := s.scs.Save(sc); err != nil {
			return false, nil, err
		}
	} else {
		logger.Infof("No sigchain for %s", kid)
	}

	if err := s.scs.Index(kid); err != nil {
		return false, nil, err
	}

	res, err := s.users.Update(ctx, kid)
	if err != nil {
		return false, nil, err
	}

	return true, res, nil
}
