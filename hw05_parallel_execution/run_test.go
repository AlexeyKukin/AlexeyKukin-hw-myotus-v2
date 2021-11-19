package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	// Дополнительные тесты
	// Тест на параллельность за метод тестирования на каналах спасибо Алексею Бакину.

	t.Run("Concurrency test on require.Eventually", func(t *testing.T) {
		var runTasksCount int32
		errCh := make(chan error)
		testCh := make(chan struct{})
		tasksCount := 5
		tasks := make([]Task, 0, tasksCount)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				<-testCh
				return nil
			})
		}
		go func() {
			errCh <- Run(tasks, tasksCount, 1)
		}()

		require.Eventually(t, func() bool { return atomic.LoadInt32(&runTasksCount) == 5 },
			time.Second, time.Millisecond, "Задачи небыли запущены параллельно! должно быть 5, запущено:%v",
			atomic.LoadInt32(&runTasksCount))
		close(testCh)

		var rErr error
		require.Eventually(t, func() bool {
			select {
			case rErr = <-errCh:
				return true
			default:
				return false
			}
		}, time.Second, time.Millisecond)

		require.NoError(t, rErr)
		close(errCh)
	})

	t.Run("Тест ошибок при m < 1. Выдается ошибка ErrErrorsLimitExceeded и таски не запускаются",
		func(t *testing.T) {
			var runNegativeTasksCount int32
			var runZerroTasksCount int32
			tasksCount := 50
			workersCount := 10
			tasksNegative := make([]Task, 0, tasksCount)
			tasksZerro := make([]Task, 0, tasksCount)
			// Заполняем пул тасок. Все таски выдают ошибки. Подсчитываем кол-во запущенных тасок
			for i := 0; i < tasksCount; i++ {
				err := fmt.Errorf("error from task %d", i)
				tasksNegative = append(tasksNegative, func() error {
					atomic.AddInt32(&runNegativeTasksCount, 1)
					return err
				})
			}
			// Заполняем пул тасок. Все таски выдают ошибки. Подсчитываем кол-во запущенных тасок
			for i := 0; i < tasksCount; i++ {
				err := fmt.Errorf("error from task %d", i)
				tasksZerro = append(tasksZerro, func() error {
					atomic.AddInt32(&runZerroTasksCount, 1)
					return err
				})
			}
			//

			maxErrorsCount := -123
			errNegative := Run(tasksNegative, workersCount, maxErrorsCount)

			maxErrorsCount = 0
			errZerro := Run(tasksZerro, workersCount, maxErrorsCount)

			require.Truef(t, (errors.Is(errNegative, ErrErrorsLimitExceeded)),
				"Негативное значение не выдает ошибку: ErrErrorsLimitExceeded Ошибка: %v", errNegative)
			require.Truef(t, (atomic.LoadInt32(&runNegativeTasksCount) == 0),
				"Негативное значение запускает задачи на исполнение:", runNegativeTasksCount)
			require.Truef(t, (errors.Is(errZerro, ErrErrorsLimitExceeded)),
				"Нулевое значение не выдает ошибку: ErrErrorsLimitExceeded Ошибка: %v", errZerro)
			require.Truef(t, (atomic.LoadInt32(&runZerroTasksCount) == 0),
				"Нулевое значение запускает задачи на исполнение:", runZerroTasksCount)
		})
}
