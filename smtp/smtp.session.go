package smtp

import (
	"crypto/subtle"
	"io"
	"net"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"xn--gckvb8fzb.com/glides/runtime"
	"xn--gckvb8fzb.com/switchyard/errs"
	"xn--gckvb8fzb.com/switchyard/services/config"
	"xn--gckvb8fzb.com/switchyard/services/dispatch"
)

type backend struct {
	rt      *runtime.Runtime
	cfg     config.SMTP
	disp    *dispatch.Dispatch
	allowed []*net.IPNet
}

func newBackend(
	rt *runtime.Runtime,
	cfg config.SMTP,
	disp *dispatch.Dispatch,
	allowed []*net.IPNet,
) *backend {
	return &backend{rt: rt, cfg: cfg, disp: disp, allowed: allowed}
}

func (b *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	remote := c.Conn().RemoteAddr()
	if !b.sourceAllowed(remote) {
		b.rt.Warn("status", "rejected source", "remote", remote.String())
		return nil, &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 7, 1},
			Message:      "Access denied",
		}
	}

	b.rt.Debug("status", "session", "remote", remote.String())
	return &session{be: b}, nil
}

func (b *backend) sourceAllowed(remote net.Addr) bool {
	if len(b.allowed) == 0 {
		return true
	}

	host, _, err := net.SplitHostPort(remote.String())
	if err != nil {
		host = remote.String()
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	for _, network := range b.allowed {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (b *backend) authEnabled() bool {
	return b.cfg.Username != ""
}

func (b *backend) checkAuth(username, password string) error {
	userOK := subtle.ConstantTimeCompare(
		[]byte(username), []byte(b.cfg.Username)) == 1
	passOK := subtle.ConstantTimeCompare(
		[]byte(password), []byte(b.cfg.Password)) == 1
	if userOK && passOK {
		return nil
	}
	return errs.ErrAuthFailed
}

type session struct {
	be   *backend
	from string
	to   []string
}

func (s *session) AuthMechanisms() []string {
	if !s.be.authEnabled() {
		return nil
	}
	return []string{sasl.Plain, sasl.Login}
}

func (s *session) Auth(mech string) (sasl.Server, error) {
	if !s.be.authEnabled() {
		return nil, smtp.ErrAuthUnsupported
	}

	switch mech {
	case sasl.Plain:
		return sasl.NewPlainServer(func(_, username, password string) error {
			return s.be.checkAuth(username, password)
		}), nil
	case sasl.Login:
		return &loginServer{authenticate: s.be.checkAuth}, nil
	default:
		return nil, smtp.ErrAuthUnsupported
	}
}

type loginServer struct {
	authenticate func(username, password string) error
	username     string
	asked        bool
}

func (a *loginServer) Next(response []byte) (challenge []byte, done bool, err error) {
	if !a.asked {
		if response == nil {
			return []byte("Username:"), false, nil
		}
		a.username = string(response)
		a.asked = true
		return []byte("Password:"), false, nil
	}
	return nil, true, a.authenticate(a.username, string(response))
}

func (s *session) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	if len(s.to) == 0 {
		return errs.ErrNoRecipients
	}

	m, err := parse(r)
	if err != nil {
		return err
	}
	m.Recipients = s.to
	if m.From == "" {
		m.From = s.from
	}

	s.be.rt.Info("status", "accepted",
		"from", m.From, "recipients", m.Recipients, "subject", m.Subject)

	return s.be.disp.Deliver(m)
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}
