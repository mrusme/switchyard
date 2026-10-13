package xmpp

import (
	"sync"

	"xn--gckvb8fzb.com/glides/helpers/xmppclient"
	"xn--gckvb8fzb.com/glides/runtime"
	"xn--gckvb8fzb.com/switchyard/errs"
	"xn--gckvb8fzb.com/switchyard/services/config"
)

type XMPP struct {
	rt   *runtime.Runtime
	opts xmppclient.Options

	mu        sync.Mutex
	client    *xmppclient.Client
	onMessage func(xmppclient.Message)

	stop     chan struct{}
	stopOnce sync.Once
}

func New(rt *runtime.Runtime) (*XMPP, error) {
	t := new(XMPP)
	t.rt = rt
	t.stop = make(chan struct{})

	cfg, err := config.GetXMPP(rt.Config())
	if err != nil {
		return nil, err
	}
	if cfg.Server == "" || cfg.Username == "" || cfg.Password == "" {
		return nil, errs.ErrXMPPServerRequired
	}

	t.opts = xmppclient.Options{
		Host:               cfg.Server,
		User:               cfg.Username,
		Password:           cfg.Password,
		DirectTLS:          cfg.DirectTLS,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Status:             cfg.Status,
	}

	return t, nil
}

func (t *XMPP) OnMessage(fn func(xmppclient.Message)) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.onMessage = fn
	if t.client != nil {
		t.client.SetOnMessage(fn)
	}
}

func (t *XMPP) Startup() error {
	t.rt.Info("status", "connecting", "host", t.opts.Host)

	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.connect(); err != nil {
		t.rt.Error("status", "connect failed, will retry on first send",
			"host", t.opts.Host, "error", err)
	}

	return nil
}

func (t *XMPP) Run() error {
	<-t.stop

	return nil
}

func (t *XMPP) Shutdown() error {
	t.rt.Info("status", "shutdown")

	t.stopOnce.Do(func() { close(t.stop) })

	t.mu.Lock()
	defer t.mu.Unlock()

	t.disconnect()

	return nil
}

func (t *XMPP) Send(to, body string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err = t.ensureConnected(); err != nil {
		return err
	}

	if err = t.client.Send(to, body); err != nil {
		t.rt.Error("status", "send failed", "to", to, "error", err)
		return err
	}

	t.rt.Debug("status", "sent", "to", to)

	return nil
}

func (t *XMPP) ensureConnected() error {
	if t.client != nil {
		select {
		case <-t.client.Done():
			t.rt.Info("status", "connection ended, reconnecting", "host", t.opts.Host)
		default:
			err := t.client.Ping()
			if err == nil {
				return nil
			}
			t.rt.Error("status", "ping failed, reconnecting", "error", err)
		}

		t.disconnect()
	}

	return t.connect()
}

func (t *XMPP) connect() (err error) {
	t.rt.Debug("status", "dialing", "host", t.opts.Host)

	opts := t.opts
	opts.OnMessage = t.onMessage

	if t.client, err = xmppclient.Dial(t.rt.Logger(), opts); err != nil {
		t.client = nil
		return err
	}

	return nil
}

func (t *XMPP) disconnect() {
	if t.client == nil {
		return
	}

	t.client.Close()
	t.client = nil
}
