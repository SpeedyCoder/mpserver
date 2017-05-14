package main

import (
	"mpserver"
	"time"
)

func main() {
	worker := mpserver.MakeComponent(
		func (job mpserver.Job) {
			time.Sleep(time.Second*5)
			job.SetResult("Hello World!")
	})

	lb := mpserver.DynamicLoadBalancer(
			worker, 10, time.Second, time.Second*60, )
	// lb := mpserver.StaticLoadBalancer(worker, 10)

	in := mpserver.GetChan()
	out := mpserver.GetChan()
	go lb(in, out)
	go mpserver.StringWriter(out)

    mpserver.Listen("/", in, nil)
    mpserver.ListenAndServe(":3000", nil)
}