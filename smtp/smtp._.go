package smtp

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"xn--gckvb8fzb.com/glides/runtime"
	"xn--gckvb8fzb.com/switchyard/services/config"
	"xn--gckvb8fzb.com/switchyard/services/dispatch"
)

type listener struct {
	srv      *smtp.Server
	addr     string
	implicit bool
}

type Server struct {
	rt        *runtime.Runtime
	cfg       config.SMTP
	listeners []listener
}

func New(rt *runtime.Runtime, disp *dispatch.Dispatch) (*Server, error) {
	s := new(Server)
	s.rt = rt

	cfg, err := config.GetSMTP(rt.Config())
	if err != nil {
		return nil, err
	}
	s.cfg = cfg

	allowed, err := parseAllowedIPs(cfg.AllowedIPs)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		rt.Warn("smtp", "no source allowlist, accepting from any address")
	}

	be := newBackend(rt, cfg, disp, allowed)

	var tlsConfig *tls.Config
	if cfg.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.Cert, cfg.TLS.Key)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	if cfg.TLS.Enable {
		s.listeners = append(s.listeners, listener{
			srv:      s.newServer(be, cfg.SMTPSAddress, tlsConfig, false),
			addr:     cfg.SMTPSAddress,
			implicit: true,
		})
	} else {
		rt.Warn("smtp", "TLS disabled, not starting implicit-TLS listener",
			"address", cfg.SMTPSAddress)
	}

	s.listeners = append(s.listeners, listener{
		srv:      s.newServer(be, cfg.SubmissionAddress, tlsConfig, !cfg.TLS.Enable),
		addr:     cfg.SubmissionAddress,
		implicit: false,
	})

	return s, nil
}

func (s *Server) newServer(
	be *backend,
	addr string,
	tlsConfig *tls.Config,
	allowInsecureAuth bool,
) *smtp.Server {
	srv := smtp.NewServer(be)
	srv.Addr = addr
	srv.Domain = s.cfg.Domain
	srv.ReadTimeout = parseDuration(s.cfg.ReadTimeout, 30*time.Second)
	srv.WriteTimeout = parseDuration(s.cfg.WriteTimeout, 30*time.Second)
	srv.MaxMessageBytes = s.cfg.MaxMessageBytes
	srv.MaxRecipients = s.cfg.MaxRecipients
	srv.TLSConfig = tlsConfig
	srv.AllowInsecureAuth = allowInsecureAuth
	return srv
}

func (s *Server) Run() error {
	errCh := make(chan error, len(s.listeners))

	for _, l := range s.listeners {
		go func(l listener) {
			s.rt.Info("status", "listening",
				"address", l.addr, "implicit_tls", l.implicit)

			var err error
			if l.implicit {
				err = l.srv.ListenAndServeTLS()
			} else {
				err = l.srv.ListenAndServe()
			}
			if err != nil {
				s.rt.Error("status", "listener stopped",
					"address", l.addr, "error", err)
			}
			errCh <- err
		}(l)
	}

	return <-errCh
}

func (s *Server) Shutdown() error {
	for _, l := range s.listeners {
		if err := l.srv.Close(); err != nil {
			s.rt.Debug("status", "close error", "address", l.addr, "error", err)
		}
	}
	return nil
}

func parseAllowedIPs(entries []string) ([]*net.IPNet, error) {
	var networks []*net.IPNet

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.Contains(entry, "/") {
			_, network, err := net.ParseCIDR(entry)
			if err != nil {
				return nil, err
			}
			networks = append(networks, network)
			continue
		}

		ip := net.ParseIP(entry)
		if ip == nil {
			return nil, &net.ParseError{Type: "IP address", Text: entry}
		}

		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		networks = append(networks, &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(bits, bits),
		})
	}

	return networks, nil
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
