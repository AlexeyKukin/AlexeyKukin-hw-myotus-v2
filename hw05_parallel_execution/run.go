package hw05parallelexecution

import (
	//	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

// Тип задача, принимает функцию, возвращает ошибку.
type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
// Функция запускает n воркеров, которые исполняют задания пока не возникает m ошибок.
func Run(tasks []Task, n, m int) error {
	// Создаем канал тасков с буфером равным их кол-ву и заполняем его ими, канал потокобезопасен и поэтому воркеры легко разберут задачи.
	chTsk := make(chan Task, len(tasks))
	defer close(chTsk)
	er := int32(m)
	for _, v := range tasks {
		chTsk <- v
	}

	wg := sync.WaitGroup{}
	//	ctx := context.Background()

	for i := 0; i <= n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		all:
			for {
				if er <= 0 {
					return
				}
				select {
				case fn := <-chTsk:
					{
						err := fn()
						if err != nil {
							atomic.AddInt32(&er, -1)
						}
					}
				default:
					{
						break all
					}

				}
			}
			//

		}()
	}

	// Place your code here.
	// Возвращаем или nil или ошибку "ErrErrorsLimitExceeded"
	wg.Wait()
	if er <= 0 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
