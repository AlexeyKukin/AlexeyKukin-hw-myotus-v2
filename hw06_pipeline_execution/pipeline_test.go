package hw06pipelineexecution

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator. Принимает строку и функцию, принимающую именованную переменную v типа пустой интерфейс,
	// И возвращающую пустой интерфейс обратно. А возвращает тип Stage.
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		// Возвращает функцию нужного нам типа Stage:
		return func(in In) Out {
			// Эта часть генерирует канал типа OUT возвращает его и наполняет в горутине результатами.
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			// Вот тут возвращаем канал.
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}
	// Простой кейс
	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}
		// Эта горутина наполняет канал in
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		// Тут мы итерируемся по каналу Out, который возвращает ExecutePipeline это позволяет нам полностью дождаться
		// пока канал out ExecutePipeline не будет закрыт
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})

	t.Run("Zerro data send", func(t *testing.T) {
		in := make(Bi)
		data := []int{}

		// Эта горутина наполняет канал in
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()
		result := make([]string, 0)
		// Тут мы итерируемся по каналу Out, который возвращает ExecutePipeline это позволяет нам полностью
		// дождаться пока канал out ExecutePipeline не будет закрыт
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}

		require.Equal(t, []string{}, result)
	})
}
