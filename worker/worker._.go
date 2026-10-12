package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	glideserrs "xn--gckvb8fzb.com/glides/errs"
	"xn--gckvb8fzb.com/glides/runtime"
	glidesconfig "xn--gckvb8fzb.com/glides/services/config"
	glidesdispatch "xn--gckvb8fzb.com/glides/services/dispatch"
	"xn--gckvb8fzb.com/switchyard/errs"
	"xn--gckvb8fzb.com/switchyard/helpers/jid"
	"xn--gckvb8fzb.com/switchyard/models/asyncjob"
	"xn--gckvb8fzb.com/switchyard/models/message"
)

type Sender interface {
	Send(to, body string) error
}

type Worker struct {
	rt       *runtime.Runtime
	sender   Sender
	redisCfg glidesconfig.Redis
	as       *asynq.Server
	asMux    *asynq.ServeMux

	stop     chan struct{}
	stopOnce sync.Once
}

func New(rt *runtime.Runtime, sender Sender) (wrk *Worker, err error) {
	wrk = new(Worker)
	wrk.rt = rt
	wrk.sender = sender
	wrk.stop = make(chan struct{})

	return wrk, nil
}

func (wrk *Worker) Startup() (err error) {
	if wrk.redisCfg, err = wrk.rt.Config().Redis(); err != nil {
		return err
	}

	wrk.as = asynq.NewServer(
		glidesdispatch.RedisConnOpt(wrk.redisCfg),
		asynq.Config{
			Logger:      wrk.rt.ALogger,
			Concurrency: wrk.redisCfg.Poolsize,
		},
	)

	wrk.asMux = asynq.NewServeMux()
	wrk.asMux.HandleFunc(glidesdispatch.TaskJob, wrk.HandleJob)

	return nil
}

func (wrk *Worker) Run() (err error) {
	if wrk.as == nil {
		return fmt.Errorf("%w: worker was not started", glideserrs.ErrProcessState)
	}

	if err = wrk.as.Start(wrk.asMux); err != nil {
		return err
	}

	<-wrk.stop

	return nil
}

func (wrk *Worker) Shutdown() error {
	wrk.stopOnce.Do(func() { close(wrk.stop) })

	if wrk.as != nil {
		wrk.as.Shutdown()
	}

	return nil
}

func (wrk *Worker) HandleJob(ctx context.Context, t *asynq.Task) (err error) {
	var job asyncjob.AsyncJob
	if err = json.Unmarshal(t.Payload(), &job); err != nil {
		return err
	}

	if job.Type != asyncjob.Delivery || job.SubType != asyncjob.XMPP {
		return fmt.Errorf("%w: %s/%s",
			errs.ErrJobPayloadInvalid, job.Type, job.SubType)
	}

	payloads, err := asyncjob.Payloads[*message.Message](job)
	if err != nil {
		return err
	}

	for _, m := range payloads {
		body := m.Body()

		for _, to := range m.Recipients {
			dest := jid.FromAddress(to)
			if err = wrk.sender.Send(dest, body); err != nil {
				return err
			}
			wrk.rt.Info("status", "delivered", "to", dest)
		}
	}

	return nil
}
