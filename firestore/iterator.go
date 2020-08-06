package firestore

import (
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/docs/events"
	"github.com/keys-pub/keys/tsutil"
	"google.golang.org/api/iterator"
)

type docsIterator struct {
	iter     *firestore.DocumentIterator
	prefix   string
	parent   string
	pathOnly bool
}

func (i *docsIterator) Next() (*docs.Document, error) {
	if i.iter == nil {
		return nil, nil
	}
	doc, err := i.iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	k := doc.Ref.ID
	if i.prefix != "" && !strings.HasPrefix(k, i.prefix) {
		// We've reached an entry not matching prefix, so end iteration
		// TODO: Is there a more efficient way to do this in the query?
		return nil, nil
	}
	kp := docs.Path(i.parent, k)

	if i.pathOnly {
		out := docs.NewDocument(kp, nil)
		out.CreatedAt = doc.CreateTime
		out.UpdatedAt = doc.UpdateTime
		return out, nil
	}

	m := doc.Data()
	b, _ := m["data"].([]byte)
	out := docs.NewDocument(kp, b)
	out.CreatedAt = doc.CreateTime
	out.UpdatedAt = doc.UpdateTime
	return out, nil
}

func (i *docsIterator) Release() {
	if i.iter != nil {
		i.iter.Stop()
	}
}

type eventIterator struct {
	iter  *firestore.DocumentIterator
	limit int64
	count int64
}

func (i *eventIterator) Next() (*events.Event, error) {
	if i.iter == nil {
		return nil, nil
	}
	doc, err := i.iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if i.limit != 0 && i.count >= i.limit {
		return nil, nil
	}
	var event events.Event
	if err := doc.DataTo(&event); err != nil {
		return nil, err
	}
	event.Timestamp = tsutil.Millis(doc.CreateTime)
	i.count++
	return &event, nil

}

func (i *eventIterator) Release() {
	if i.iter != nil {
		i.iter.Stop()
	}
}
