package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	/*
		Канал in поступает нам на вход и содержит данные для обработки
		Канал done у нас тоже передается параметром внутрь функции, при его закрытии, мы должны завершить работу.
		Cтейджи передаются как []Stage
		Нам необходимо вернуть канал Out, который нам нужно создать внутри нашей функции и закрыть по завершении работы

	*/
	out := make(Bi)

	// Тут мы наконец то возвращаем наш канал из функции. В него мы будем передавать результат работы пайплайна
	return out
}
