package sdb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/keys-pub/keys/ds"
)

const incrementBatchSize = 1000

func (d *DB) increment(ctx context.Context, path string) error {
	d.rwmtx.Lock()
	defer d.rwmtx.Unlock()
	doc, err := d.get(ctx, ds.Path(path))
	if err != nil {
		return err
	}

	var inc int
	if doc == nil {
		inc = 1
	} else {
		i, err := strconv.Atoi(string(doc.Data))
		if err != nil {
			return err
		}
		inc = i
	}

	if err := d.set(ctx, ds.Path(path), []byte(strconv.Itoa(inc+incrementBatchSize))); err != nil {
		return err
	}

	logger.Debugf("Setting increment batch: %d", inc)
	d.inc = inc
	d.incMax = inc + incrementBatchSize - 1

	return nil
}

// Increment returns the current increment as an orderable string that persists
// across opens.
// => 000000000000001, 000000000000002 ...
// This is batched. When the increment runs out for the current batch, it
// gets a new batch.
// The increment value is saved in the database at the specified path.
// There may be large gaps between increments (of batch size) after re-opens.
func (d *DB) Increment(ctx context.Context, path string) (string, error) {
	if d.inc == 0 || d.inc >= d.incMax {
		if err := d.increment(ctx, path); err != nil {
			return "", err
		}
	}
	d.inc++
	if d.inc > 999999999999999 {
		panic("increment too large")
	}
	return fmt.Sprintf("%015d", d.inc), nil
}
