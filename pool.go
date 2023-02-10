package go_worker_pool

import (
	"context"
	"fmt"
	"sync"
)

const defaultPoolSize = 3

var (
	once                    sync.Once
	pool                    *workerPool
	invalidRequiredParamErr = fmt.Errorf("invalid required parameter")
)

type WorkerPool interface {
	AddTasks(tasks []Executor)
	GetResults() chan *TaskResult
}

type workerPool struct {
	poolSize int
	results  chan *TaskResult
	tasks    chan Executor
}

func NewWorkerPool(ctx context.Context, size int) WorkerPool {
	once.Do(func() {
		if size == 0 {
			size = defaultPoolSize
		}
		pool = &workerPool{
			poolSize: size,
			results:  make(chan *TaskResult, size),
			tasks:    make(chan Executor, size),
		}
		go pool.run(ctx)
	})
	return pool
}

func (wp *workerPool) run(ctx context.Context) {
	defer close(pool.results)
	var wg sync.WaitGroup
	for i := 0; i < pool.poolSize; i++ {
		wg.Add(1)
		go pool.worker(ctx, &wg, pool.tasks, pool.results)
	}
	wg.Wait()
}

func (wp *workerPool) AddTasks(inTasks []Executor) {
	defer close(wp.tasks)
	for i := range inTasks {
		wp.tasks <- inTasks[i]
	}
}

func (wp *workerPool) GetResults() chan *TaskResult {
	return wp.results
}

func (wp *workerPool) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan Executor, results chan<- *TaskResult) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			results <- &TaskResult{
				Err: ctx.Err(),
			}
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}
			results <- task.Execute(ctx)
		}
	}
}

type Executor interface {
	Execute(ctx context.Context) *TaskResult
}

type TaskResult struct {
	TaskID string
	Value  any
	Err    error
}

type task struct {
	taskID string
	execFn ExecutionFn
	args   any
}

func NewTask(id int, fn ExecutionFn, args any) (Executor, error) {
	if fn == nil {
		return nil, invalidRequiredParamErr
	}
	return &task{
		taskID: fmt.Sprintf("task-%d", id),
		execFn: fn,
		args:   args,
	}, nil
}

func (t *task) Execute(ctx context.Context) *TaskResult {
	val, err := t.execFn(ctx, t.args)
	if err != nil {
		return &TaskResult{
			TaskID: t.taskID,
			Err:    err,
		}
	}
	return &TaskResult{
		TaskID: t.taskID,
		Value:  val,
	}
}

type ExecutionFn func(ctx context.Context, args any) (any, error)
