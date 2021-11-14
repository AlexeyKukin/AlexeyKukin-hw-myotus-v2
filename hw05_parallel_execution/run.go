package hw05parallelexecution

import (
	//	"context"
	"errors"
	"sync"
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
	for _, v := range tasks {
		chTsk <- v
	}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	//	ctx := context.Background()

	for i := 0; i <= n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		all:
			for {
				mu.Lock()
				if m <= 0 {
					mu.Unlock()
					return
				}
				mu.Unlock()

				select {
				case fn := <-chTsk:
					{
						err := fn()
						if err != nil {
							mu.Lock()
							m--
							mu.Unlock()
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
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
