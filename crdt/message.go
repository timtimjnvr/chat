package crdt

type (
	message struct {
		sender, content string
	}

	Message interface {
		GetSender() string
		GetContent() string
		UpdateContent(content string) string
	}
)

func (m *message) GetSender() string {
	return m.sender
}

func (m *message) GetContent() string {
	return m.content
}

func (m *message) UpdateContent(content string) {
	m.content = content
}
