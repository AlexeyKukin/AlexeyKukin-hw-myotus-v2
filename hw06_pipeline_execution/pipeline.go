package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in
	for _, i := range stages {

		out = ifDone(i(out), done)
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
				if ok != true {
					close(out)
					return
				}
				out <- a
			}
		}
	}()
	return out
}
