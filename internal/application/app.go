package application

import (
	"fmt"
	"github.com/xlkness/lkit-go/internal/joymicro/joyservice"
	"github.com/xlkness/lkit-go/internal/log"
	"github.com/xlkness/lkit-go/internal/web/engine"
)

// Task 不会永久执行的任务，串行用于启动前初始化或者启动后初始化工作，返回error就停止application
type Task func() error

// Worker 永久执行的工作协程，一旦停止就停止application
type Worker func() error

// Job 不会永久执行的任务，且不关心执行结果，不关心执行顺序，例如内存预热等
type Job func()

type pair struct {
	desc string
	item interface{}
}

type App struct {
	bootFlag        interface{}
	initializeTasks []pair // 启动服务前串行执行初始化任务的job
	services        []pair // rpc服务
	servers         []pair // web服务
	postRunTasks    []pair // 启动后串行执行的job
	postRunWorkers  []pair // 启动后后台永久执行的工作协程，一旦推出就停止application
	parallelJobs    []pair // 启动services、servers后并行执行的任务，不关心结果，例如内存数据的预热等
}

func NewApp(options ...AppOption) *App {
	app := new(App)
	app.applyOptions(options...)
	return app
}

// WithInitializeTask app完成init之后run之前执行的任务，可以用来初始化某些业务或者检查配置等
func (app *App) WithInitializeTask(desc string, task Task) *App {
	app.initializeTasks = append(app.initializeTasks, pair{desc, task})
	return app
}

// WithServer 添加web服务器
func (app *App) WithServer(desc string, server *engine.Engine) *App {
	app.servers = append(app.servers, pair{desc, server})
	return app
}

// WithService 添加rpc服务
func (app *App) WithService(desc string, service *joyservice.ServicesManager) *App {
	app.servers = append(app.servers, pair{desc, service})
	return app
}

// WithPostTask app run之后执行的任务，一般做临时检查任务，可以用来服务启动后加载数据检查等
func (app *App) WithPostTask(desc string, task Task) *App {
	app.postRunTasks = append(app.postRunTasks, pair{desc, task})
	return app
}

// WithPostWorker 完成post task之后执行的后台任务，报错退出等app也会退出，一般做永久的关键后台逻辑
func (app *App) WithPostWorker(desc string, worker Worker) *App {
	app.postRunWorkers = append(app.postRunWorkers, pair{desc, worker})
	return app
}

// WithParallelJob 完成post task之后执行的并行后台任务，一般做永久的不关键后台逻辑，例如内存预热等
func (app *App) WithParallelJob(desc string, job Job) *App {
	app.parallelJobs = append(app.parallelJobs, pair{desc, job})
	return app
}

func (app *App) applyOptions(options ...AppOption) *App {
	for _, option := range options {
		option.Apply(app)
	}
	return app
}

func (app *App) run() (err error) {
	waitChan := make(chan error, 1)

	// 启动前的初始化任务
	for _, j := range app.initializeTasks {
		curErr := j.item.(Task)()
		if curErr != nil {
			err = fmt.Errorf("run initialize task(%s) return error:%v", j.desc, curErr)
			return
		}
	}

	// 启动rpc服务
	for _, pair := range app.services {
		go func(desc string, s *joyservice.ServicesManager) {
			curErr := s.Run()
			if curErr != nil {
				waitChan <- fmt.Errorf("service %s run on %v error:%v", desc, s.Addr, err)
			}
		}(pair.desc, pair.item.(*joyservice.ServicesManager))
	}

	defer app.stopServices()

	// 启动web服务
	for _, pair := range app.servers {
		go func(desc string, s *engine.Engine) {
			err := s.Run()
			if err != nil {
				waitChan <- fmt.Errorf("server %s error:%v", desc, err)
			}
		}(pair.desc, pair.item.(*engine.Engine))
	}

	defer app.stopServers()

	// 启动后串行执行的job
	for _, j := range app.postRunTasks {
		curErr := j.item.(Task)()
		if curErr != nil {
			err = fmt.Errorf("run post task %s return error:%v", j.desc, curErr)
			return
		}
	}

	// 启动后串行执行的工作协程
	for _, pair := range app.postRunWorkers {
		go func(desc string, g Worker) {
			curErr := g()
			if curErr != nil {
				waitChan <- fmt.Errorf("run post worker %s return error:%v", desc, curErr)
			}
		}(pair.desc, pair.item.(Worker))
	}

	// 启动后的并行job
	for _, j := range app.parallelJobs {
		go j.item.(Job)()
	}

	select {
	case anyErr := <-waitChan:
		log.Critif("scheduler stop with execute error:%v", anyErr)
		return anyErr
	}
}

func (app *App) stop() {
	app.stopServices()
	app.stopServers()
}

func (app *App) stopServers() {
	for _, s := range app.servers {
		s.item.(*engine.Engine).Stop()
	}
}

func (app *App) stopServices() {
	for _, s := range app.services {
		s.item.(*joyservice.ServicesManager).Stop()
	}
}
