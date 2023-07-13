package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/worker"
)

type MyJob struct {
}

func (m MyJob) Execute(_ context.Context, i int64) (int64, error) {
	fmt.Println("EXECUTING JOB EREEE")
	return 69, nil
}

func main() {

	a := make(chan worker.Executor)
	b := make(chan worker.Resulter)

	x := worker.NewWorker(a, b)

	go func() {
		if err := x.Start(); err != nil {
			os.Exit(1)
		}
	}()

	a <- job.Executor[int64, int64]{Job: MyJob{}}

	time.Sleep((time.Second * 1))

	z := <-b
	fmt.Println(z.Result())
	time.Sleep((time.Second * 10000))

}
