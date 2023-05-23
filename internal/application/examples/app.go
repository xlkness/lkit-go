package main

import (
	"fmt"
	"github.com/xlkness/lkit-go/internal/application"
	"github.com/xlkness/lkit-go/internal/log"
	"time"
)

type Scheduler1Flag struct {
	F1 string `name:"id" desc:"boot id" value:"default value"`
}

var flag = &Scheduler1Flag{}

func main() {
	app := application.NewApp(application.WithAppBootFlag(flag))
	app.WithInitializeTask("check pre task", func() error {
		fmt.Printf("check pre task\n")
		return nil
	})
	app.WithPostTask("check post task", func() error {
		fmt.Printf("check post task\n")
		return nil
	})
	app.WithPostWorker("bg worker", func() error {
		fmt.Printf("start bg worker\n")
		time.Sleep(time.Second * 2)
		return fmt.Errorf("exit bg worker")
	})
	app.WithParallelJob("bg job1", func() {
		fmt.Printf("bg job1, common flag:%+v\n", application.CommonBootFlag)
		return
	})
	app.WithParallelJob("bg job2", func() {
		fmt.Printf("bg job2, flag:%v\n", flag.F1)
		return
	})
	app1 := application.NewApp()
	app1.WithPostTask("app1 post task", func() error {
		log.Infof("app1 post task")
		return nil
	})

	scd := application.NewScheduler()
	scd.WithApp("app1", app)
	scd.WithApp("app2", app1)
	err := scd.Run()
	if err != nil {
		panic(err)
	}
}
