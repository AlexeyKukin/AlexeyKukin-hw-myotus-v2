package hw06pipelineexecution

import (
	"strconv"
	"time"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in
	if len(stages) == 0 {
		out = ifDone(myDSS(out), done)
	} else {
		for _, i := range stages {
			out = ifDone(i(out), done)
		}
	}
	return out
}

func ifDone(in In, done In) Out {
	out := make(Bi)
	go func() {
		for {
			select {
			case <-done:
				{
					close(out)
					return
				}
			case a, ok := <-in:
				if !ok {
					close(out)
					return
				}
				out <- a
				//			default:
			}
		}
	}()
	return out
}

func myDSS(in In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for v := range in {
			time.Sleep(time.Millisecond * 100)
			out <- strconv.Itoa(v.(int))
		}
	}()
	return out
}
