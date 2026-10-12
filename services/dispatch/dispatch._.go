package dispatch

import (
	"fmt"

	glideserrs "xn--gckvb8fzb.com/glides/errs"
	"xn--gckvb8fzb.com/glides/runtime"
	glidesdispatch "xn--gckvb8fzb.com/glides/services/dispatch"
)

type Dispatch struct {
	*glidesdispatch.Dispatch

	rt *runtime.Runtime
}

func New(rt *runtime.Runtime) (disp *Dispatch, err error) {
	disp = new(Dispatch)

	disp.rt = rt
	if disp.Dispatch = glidesdispatch.From(rt); disp.Dispatch == nil {
		return nil, fmt.Errorf("%w: %s", glideserrs.ErrServiceNotFound, glidesdispatch.ServiceID)
	}

	return disp, nil
}
