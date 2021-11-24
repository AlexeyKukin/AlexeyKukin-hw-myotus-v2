package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

// Тип задача, принимает функцию, возвращает ошибку.
type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// Создаем канал типа Task с буфером равным их кол-ву и заполняем его ими, канал потокобезопасен и поэтому
	// воркеры легко разберут задачи из него не мешая друг другу.
	chTsk := make(chan Task, len(tasks))
	defer close(chTsk)
	for _, v := range tasks {
		chTsk <- v
	}
	// Для работы нам понадобится Mutex и Wait группа из пакета Sync.
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	// Делаем нужное количество воркеров
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
				// Воркер либо берет задание из канала, либо завершает свое существование.
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
		}()
	}
	// Ожидаем пока все горутины остановятся из-за отсутствия заданий, или из-за превышения лимита ошибок.
	wg.Wait()
	// Возвращаем или nil или ошибку "ErrErrorsLimitExceeded"
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
