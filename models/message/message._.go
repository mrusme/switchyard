package message

import "strings"

type Message struct {
	From       string   `json:"from"`
	Subject    string   `json:"subject"`
	Text       string   `json:"text"`
	Recipients []string `json:"recipients"`
}

func New() *Message {
	return new(Message)
}

func (m *Message) Body() string {
	var b strings.Builder

	if m.From != "" {
		b.WriteString("From: ")
		b.WriteString(m.From)
		b.WriteString("\n")
	}
	if m.Subject != "" {
		b.WriteString("Subject: ")
		b.WriteString(m.Subject)
		b.WriteString("\n")
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	b.WriteString(m.Text)

	return b.String()
}
