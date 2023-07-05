package lkit_go

import (
	"github.com/xlkness/lkit-go/internal/application"
)

type Scheduler = application.Scheduler
type SchedulerOption = application.SchedulerOption

// ApplicationDescInfo 调度器创建app依赖的描述信息
type ApplicationDescInfo = application.ApplicationDescInfo
type Application = application.Application
type AppOption = application.AppOption

// CommBootFlag 调度器的全局通用启动参数
type CommBootFlag = application.CommBootFlag

func NewApplicationDescInfo(name string, initFunc func(globalBootFlag *CommBootFlag, globalBootFile interface{}, app *Application) error, options ...AppOption) *ApplicationDescInfo {
	adi := application.NewApplicationDescInfo(name, initFunc)
	adi.WithOptions(options...)
	return adi
}

// NewScheduler 创建调度器，调度器可以创建多个app和运行app
func NewScheduler(options ...SchedulerOption) *Scheduler {
	return application.NewScheduler(options...)
}

// WithAppBootFlag app注入一个启动参数结构体
func WithAppBootFlag(flag interface{}) AppOption {
	return application.WithAppBootFlag(flag)
}

// WithSchedulerBootConfigFileContent 设置启动配置文件的解析结构，指针结构体类型，不设置默认无起服配置，默认以yaml解析
func WithSchedulerBootConfigFileContent(content interface{}) SchedulerOption {
	return application.WithSchedulerBootConfigFileContent(content)
}

// WithAppBootConfigFileParser 设置起服文件解析函数，默认yaml格式
func WithAppBootConfigFileParser(f func(content []byte, out interface{}) error) SchedulerOption {
	return application.WithSchedulerBootConfigFileParser(f)
}

func WithSchedulerLogFileLevel(level LogLevel) SchedulerOption {
	return application.WithSchedulerLogFileLevel(level)
}

// WithSchedulerCreateOneAppOption 通过scheduler创建一个app
func WithSchedulerCreateOneAppOption(appDescInfo *ApplicationDescInfo) SchedulerOption {
	return application.WithSchedulerCreateOneAppOption(appDescInfo)
}
