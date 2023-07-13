package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/worker"
)

type MyJob struct {
}

func (m MyJob) Execute(_ context.Context, i int64) (int64, error) {
	fmt.Println("EXECUTING JOB EREEE")
	return 69, nil
}

func main() {

	// a := make(chan worker.Executor)
	// b := make(chan worker.Resulter)

	x := worker.NewPool(1, log.NewLogger(os.Stdout, "thread-pool"))

	x.Start()
	for i := 0; i < 100; i++ {
		x.AddTask(job.Executor[int64, int64]{Job: MyJob{}})
	}
	time.Sleep((time.Second * 1))

	for {
		select {
		case res := <-x.RespChan():
			fmt.Println(res.Result())
		case _ = <-time.After(time.Second * 5):
			x.Stop()
			return
		}
	}
}
