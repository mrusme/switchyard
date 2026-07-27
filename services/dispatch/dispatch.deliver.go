package dispatch

import (
	"time"

	"github.com/hibiken/asynq"
	"xn--gckvb8fzb.com/switchyard/models/asyncjob"
	"xn--gckvb8fzb.com/switchyard/models/message"
)

const (
	DELIVER_MAX_RETRY int           = 10
	DELIVER_TIMEOUT   time.Duration = 5 * time.Minute
)

func (disp *Dispatch) Deliver(m *message.Message) (err error) {
	j := asyncjob.New("", asyncjob.Delivery, asyncjob.XMPP)
	if err = j.SetPayload(m); err != nil {
		return err
	}

	if err = disp.Job(j,
		asynq.MaxRetry(DELIVER_MAX_RETRY),
		asynq.Timeout(DELIVER_TIMEOUT),
	); err != nil {
		return err
	}

	disp.rt.Debug("status", "enqueued",
		"id", j.ID, "recipients", m.Recipients)

	return nil
}
