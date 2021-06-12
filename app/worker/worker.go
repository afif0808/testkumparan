package worker

type Worker struct {
	jobChannel chan func()
	workerPool chan chan func()
}

func NewWorker(workerPool chan chan func()) *Worker {
	return &Worker{
		workerPool: workerPool,
		jobChannel: make(chan func()),
	}
}

func (w Worker) Start() {
	go func() {
		for {
			w.workerPool <- w.jobChannel
			job := <-w.jobChannel
			job()
		}
	}()
}

func NewDispatcher(maxWorker int) *Dispatcher {
	return &Dispatcher{
		maxWorker:  maxWorker,
		JobQueue:   make(chan func()),
		workerPool: make(chan chan func(), maxWorker),
	}
}

type Dispatcher struct {
	maxWorker  int
	workerPool chan chan func()
	JobQueue   chan func()
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorker; i++ {
		w := NewWorker(d.workerPool)
		w.Start()
	}
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for job := range d.JobQueue {
		go func(job func()) {
			jobChannel := <-d.workerPool
			jobChannel <- job
		}(job)
	}
}
