package lkit_go

import "github.com/xlkness/lkit-go/internal/application"

type Scheduler = application.App
type SchedulerOption = application.AppOption
type Application = application.Scheduler
type AppOption = application.SchedulerOption

func NewApp(options ...SchedulerOption) *Scheduler {
	return application.NewApp(options...)
}

func NewScheduler(appOptions ...AppOption) *Application {
	return application.NewScheduler(appOptions...)
}
