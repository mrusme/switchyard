package xmpp

import (
	"crypto/tls"
	"strings"
	"sync"

	goxmpp "github.com/xmppo/go-xmpp"
	"xn--gckvb8fzb.com/glides/runtime"
	"xn--gckvb8fzb.com/switchyard/errs"
	"xn--gckvb8fzb.com/switchyard/services/config"
)

type XMPP struct {
	rt   *runtime.Runtime
	opts goxmpp.Options

	mu     sync.Mutex
	client *goxmpp.Client
}

func New(rt *runtime.Runtime) (*XMPP, error) {
	t := new(XMPP)
	t.rt = rt

	cfg, err := config.GetXMPP(rt.Config())
	if err != nil {
		return nil, err
	}
	if cfg.Server == "" || cfg.Username == "" || cfg.Password == "" {
		return nil, errs.ErrXMPPServerRequired
	}

	status := cfg.Status
	if status == "" {
		status = "chat"
	}

	t.opts = goxmpp.Options{
		Host:     cfg.Server,
		User:     cfg.Username,
		Password: cfg.Password,
		NoTLS:    !cfg.DirectTLS,
		StartTLS: !cfg.DirectTLS,
		TLSConfig: &tls.Config{
			ServerName:         strings.Split(cfg.Server, ":")[0],
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
		Session:             true,
		Status:              status,
		PeriodicServerPings: true,
	}

	return t, nil
}

func (t *XMPP) Run() error {
	t.rt.Info("status", "connecting", "host", t.opts.Host)

	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.reconnect(); err != nil {
		t.rt.Error("status", "connect failed, will retry on first send",
			"host", t.opts.Host, "error", err)
	}

	return nil
}

func (t *XMPP) Shutdown() error {
	t.rt.Info("status", "shutdown")

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.client == nil {
		return nil
	}

	return t.disconnect()
}

func (t *XMPP) Send(to, body string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err = t.ensureConnected(); err != nil {
		return err
	}

	if _, err = t.client.Send(goxmpp.Chat{
		Remote: to,
		Type:   "chat",
		Text:   body,
	}); err != nil {
		t.rt.Error("status", "send failed", "to", to, "error", err)
		return err
	}

	t.rt.Debug("status", "sent", "to", to)

	return nil
}

func (t *XMPP) ensureConnected() error {
	if t.client == nil {
		return t.reconnect()
	}

	if _, err := t.client.SendKeepAlive(); err != nil {
		t.rt.Error("status", "keepalive failed, reconnecting", "error", err)
		return t.reconnect()
	}

	return nil
}

func (t *XMPP) reconnect() (err error) {
	if t.client != nil {
		if err = t.disconnect(); err != nil {
			return err
		}
	}

	t.rt.Debug("status", "dialing", "host", t.opts.Host)

	t.client, err = t.opts.NewClient()
	if err != nil {
		t.client = nil
		return err
	}

	return nil
}

func (t *XMPP) disconnect() error {
	if err := t.client.Close(); err != nil {
		t.rt.Debug("status", "close error", "error", err)
	}
	t.client = nil

	return nil
}
