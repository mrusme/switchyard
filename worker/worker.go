package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"xn--gckvb8fzb.com/switchyard/dispatch"
	"xn--gckvb8fzb.com/switchyard/helpers/jid"
	"xn--gckvb8fzb.com/switchyard/models/config"
	"xn--gckvb8fzb.com/switchyard/models/message"
	"xn--gckvb8fzb.com/switchyard/runtime"
)

type Sender interface {
	Send(to, body string) error
}

type Worker struct {
	rt     *runtime.Runtime
	sender Sender
	cfg    config.Redis
	server *asynq.Server
}

func New(rt *runtime.Runtime, sender Sender) (wrk *Worker, err error) {
	wrk = new(Worker)
	wrk.rt = rt
	wrk.sender = sender

	if wrk.cfg, err = rt.Config.Redis(); err != nil {
		return nil, err
	}

	return wrk, nil
}

func (wrk *Worker) Run() error {
	wrk.server = asynq.NewServer(
		dispatch.RedisConnOpt(wrk.cfg),
		asynq.Config{
			Logger:      wrk.rt.Logger.Asynq(),
			Concurrency: wrk.cfg.Poolsize,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(dispatch.TaskDeliver, wrk.handleDeliver)

	return wrk.server.Run(mux)
}

func (wrk *Worker) Shutdown() error {
	if wrk.server != nil {
		wrk.server.Shutdown()
	}
	return nil
}

func (wrk *Worker) handleDeliver(ctx context.Context, task *asynq.Task) error {
	var m message.Message
	if err := json.Unmarshal(task.Payload(), &m); err != nil {
		return err
	}

	body := m.Body()

	for _, to := range m.Recipients {
		dest := jid.FromAddress(to)
		if err := wrk.sender.Send(dest, body); err != nil {
			return err
		}
		wrk.rt.Info("delivered", "to", dest)
	}

	return nil
}
