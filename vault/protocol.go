package vault

import (
	"strings"

	"github.com/keys-pub/keys/ds"
)

type entity string

const (
	itemEntity      entity = "item"
	authEntity      entity = "auth"
	provisionEntity entity = "provision"
	pendingEntity   entity = "pending"
	historyEntity   entity = "history"
	configEntity    entity = "config"
)

// protocol describes the format for paths/ids.
type protocol interface {
	Path(ent entity, paths ...string) string
}

type v2 struct{}

// V2 is the version 2 protocol.
func V2() Option {
	return func(o *Options) {
		o.protocol = v2{}
	}
}

// Path for entity.
func (v v2) Path(ent entity, paths ...string) string {
	return ds.Path(string(ent), paths)
}

type v1 struct{}

// V1 is the version 1 protocol.
func V1() Option {
	return func(o *Options) {
		o.protocol = v1{}
	}
}

// Path for entity.
func (v v1) Path(ent entity, paths ...string) string {
	id := strings.Join(paths, "-")
	switch ent {
	case itemEntity:
		return id
	case authEntity:
		if id == "" {
			return "#auth"
		}
		return "#auth-" + id
	case provisionEntity:
		return "#provision-" + id
	case pendingEntity:
		return "#pending-" + id
	case historyEntity:
		return "#history-" + id
	case configEntity:
		return "#" + id
	default:
		panic("invalid entity")
	}
}
