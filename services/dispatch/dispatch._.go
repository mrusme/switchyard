package dispatch

import (
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
	disp.Dispatch = rt.Dispatch()

	return disp, nil
}
