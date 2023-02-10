## Worker Pool
#### Simple implementation of Worker Pool (Thread Pool) concurrency pattern in Golang

#### Public WorkerPool API:
````
type WorkerPool interface {
    AddTasks(tasks []Executor)
    GetResults() chan *TaskResult
}
````

#### Public Task API:
````
type Executor interface {
    Execute(ctx context.Context) *TaskResult
}
````
#### Example:
````
package main


const (
    maxWorkers = 10
    amountOfTasks = 100
)

func main() {
    ctx := context.Background()
    // start pool
    pool := NewWorkerPool(ctx, maxWorkers)
    
    // add new tasks
    go pool.AddTasks(generator())
    
    // handle result
    for res := range pool.GetResults() {
        fmt.Println(res.Value)
    }
}

func generator() []ds.Executor {
    res := make([]ds.Executor, amountOfTasks)
    for i := 0; i < amountOfTasks; i++ {
        task, _ := ds.NewTask(i, sum, []int{i, i * i})
        res[i] = task
    }
    return res
}

// simple function for sum which
// return sum of 2 values
func sum(_ context.Context, in any) (any, error) {
    val, ok := in.([]int)
    if !ok {
        return nil, fmt.Errorf("cast error")
    }
    return val[0] + val[1], nil
}
````