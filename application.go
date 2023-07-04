package lkit_go

import (
	"github.com/xlkness/lkit-go/internal/application"
	"github.com/xlkness/lkit-go/internal/log"
)

type Scheduler = application.Scheduler
type SchedulerOption = application.SchedulerOption
type Application = application.App
type AppOption = application.AppOption

func NewApp(name string, options ...AppOption) *Application {
	return application.NewApp(name, options...)
}

func NewScheduler(options ...SchedulerOption) *Scheduler {
	return application.NewScheduler(options...)
}

func WithAppBootFlag(flag interface{}) AppOption {
	return application.WithAppBootFlag(flag)
}

// WithSchedulerBootConfigFileContent 设置启动配置文件的解析结构，不设置默认无起服配置，默认以yaml解析
func WithSchedulerBootConfigFileContent(content interface{}) SchedulerOption {
	return application.WithSchedulerBootConfigFileContent(content)
}

// WithAppBootConfigFileParser 设置起服文件解析函数，默认yaml格式
func WithAppBootConfigFileParser(f func(content []byte, out interface{}) error) SchedulerOption {
	return application.WithAppBootConfigFileParser(f)
}

func WithSchedulerLogFileLevel(level log.LogLevel) SchedulerOption {
	return application.WithSchedulerLogFileLevel(level)
}
