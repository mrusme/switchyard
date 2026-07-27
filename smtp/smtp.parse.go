package smtp

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/emersion/go-message/mail"
	"xn--gckvb8fzb.com/switchyard/models/message"
)

var (
	reTag   = regexp.MustCompile(`(?s)<[^>]*>`)
	reSpace = regexp.MustCompile(`[ \t]*\n[ \t]*\n[ \t]*(\n)+`)
)

func parse(r io.Reader) (*message.Message, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	m := message.New()

	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		m.Text = strings.TrimSpace(string(raw))
		return m, nil
	}

	if subject, err := mr.Header.Subject(); err == nil {
		m.Subject = subject
	}
	if from, err := mr.Header.AddressList("From"); err == nil && len(from) > 0 {
		m.From = formatAddress(from[0])
	}

	var plain, htm string
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			break
		}

		h, ok := part.Header.(*mail.InlineHeader)
		if !ok {
			continue
		}

		ct, _, _ := h.ContentType()
		body, _ := io.ReadAll(part.Body)

		switch ct {
		case "text/plain":
			if plain == "" {
				plain = string(body)
			}
		case "text/html":
			if htm == "" {
				htm = string(body)
			}
		}
	}

	switch {
	case plain != "":
		m.Text = strings.TrimSpace(plain)
	case htm != "":
		m.Text = stripHTML(htm)
	}

	return m, nil
}

func formatAddress(a *mail.Address) string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s>", a.Name, a.Address)
	}
	return a.Address
}

func stripHTML(s string) string {
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = reSpace.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
