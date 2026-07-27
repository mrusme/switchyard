package asyncjob

import (
	glidesasyncjob "xn--gckvb8fzb.com/glides/models/asyncjob"
)

type (
	AsyncJob   = glidesasyncjob.AsyncJob
	JobType    = glidesasyncjob.JobType
	JobSubType = glidesasyncjob.JobSubType
)

const (
	Delivery JobType = "delivery"
)

const (
	XMPP JobSubType = "xmpp"
)

func New(
	targetID string,
	jobType JobType,
	jobSubType JobSubType,
) *AsyncJob {
	return glidesasyncjob.New(targetID, jobType, jobSubType)
}

func Payloads[P any](j AsyncJob) ([]P, error) {
	return glidesasyncjob.Payloads[P](j)
}
