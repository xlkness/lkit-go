package application

import "github.com/xlkness/lkit-go/internal/log"

// WithSchedulerBootConfigFileContent 设置启动配置文件的解析结构，不设置默认无起服配置，默认以yaml解析
func WithSchedulerBootConfigFileContent(content interface{}) SchedulerOption {
	return appOptionFun(func(scd *Scheduler) {
		scd.globalBootConfigFileContent = content
	})
}

// WithAppBootConfigFileParser 设置起服文件解析函数，默认yaml格式
func WithAppBootConfigFileParser(f func(content []byte, out interface{}) error) SchedulerOption {
	return appOptionFun(func(scd *Scheduler) {
		scd.globalBootConfigParser = f
	})
}

// WithSchedulerLogFileTimestampFormat 设置日志文件默认时间戳格式，默认"20060102"
//func WithSchedulerLogFileTimestampFormat(format string) SchedulerOption {
//	return appOptionFun(func(scd *Scheduler) {
//		scd.log.logFileTsFormat = format
//	})
//}

func WithSchedulerLogFileLevel(level log.LogLevel) SchedulerOption {
	return appOptionFun(func(scd *Scheduler) {
		scd.defaultLogLevel = level
	})
}

type SchedulerOption interface {
	Apply(app *Scheduler)
}

type appOptionFun func(scd *Scheduler)

func (of appOptionFun) Apply(scd *Scheduler) {
	of(scd)
}
