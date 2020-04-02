package api

// Message ...
type Message struct {
	Data []byte `json:"data"`
	ID   string `json:"id"`
}

// CreateMessageResponse ...
type CreateMessageResponse struct {
	ID string `json:"id"`
}

// MessagesResponse is the response from messages.
type MessagesResponse struct {
	Messages []*Message          `json:"messages"`
	Metadata map[string]Metadata `json:"md,omitempty"`
	Version  string              `json:"version"`
}

// MetadataFor returns metadata for Message.
func (r MessagesResponse) MetadataFor(msg *Message) Metadata {
	md, ok := r.Metadata[msg.ID]
	if !ok {
		return Metadata{}
	}
	return md
}
