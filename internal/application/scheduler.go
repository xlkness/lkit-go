package application

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/xlkness/lkit-go/internal/flags"
	"github.com/xlkness/lkit-go/internal/libsyscal"
	"github.com/xlkness/lkit-go/internal/log"
	"github.com/xlkness/lkit-go/internal/log/handler"
	"github.com/xlkness/lkit-go/internal/trace/holmes"
	"github.com/xlkness/lkit-go/internal/trace/prom"
	"github.com/xlkness/lkit-go/internal/utils"
	"github.com/xlkness/lkit-go/internal/web/engine"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Scheduler 调度器，调度多个app
type Scheduler struct {
	globalBootFlag              *CommBootFlag                          // app全局启动参数
	globalBootConfigFileContent interface{}                            // app全局配置文件内容结构体指针，为空没有配置文件解析
	globalBootConfigParser      func(in []byte, out interface{}) error // 配置文件解析函数，默认yaml
	defaultLogLevel             log.LogLevel                           // 默认debug日志等级，优先用globalBootFlag指定的日志等级
	adis                        []*ApplicationDescInfo
	apps                        []*Application // 可绑定多个app
	server                      *engine.Engine // app全局的web服务，当前暂时一个，为prometheus、pprof共用
}

// NewScheduler
func NewScheduler(appOptions ...SchedulerOption) *Scheduler {
	scd := new(Scheduler)
	scd.globalBootFlag = CommonBootFlag
	scd.applyOptions(appOptions...)
	return scd
}

func (scd *Scheduler) CreateApp(appDescInfoList ...*ApplicationDescInfo) *Scheduler {
	for _, adi := range appDescInfoList {
		scd.withAppDescInfo(adi)
	}
	return scd
}

func (scd *Scheduler) Run() error {
	// 初始化app
	err := scd.initialize()
	if err != nil {
		return err
	}

	err = scd.initApps()
	if err != nil {
		return err
	}

	// 启动调度器
	type waitInfo struct {
		desc string
		scd  *Application
		err  error
	}
	waitChan := make(chan waitInfo, 1)

	// 运行trace server
	go func() {
		log.Noticef("trace server listen on:%v", scd.server.Addr)
		err := scd.server.Run()
		if err != nil {
			waitChan <- waitInfo{"trace server", nil, err}
		}
	}()

	for _, app := range scd.apps {
		go func(app *Application) {
			err := app.run()
			if err != nil {
				// 返回调度器的报错
				waitChan <- waitInfo{app.Name, app, err}
			}
		}(app)
	}

	watchSignChan := libsyscal.WatchSignal1()

	defer scd.Stop()

	select {
	case signal := <-watchSignChan:
		log.Noticef("Application receive signal(%v), will graceful stop", signal)
		return nil
	case errInfo := <-waitChan:
		err := fmt.Errorf("Application receive scheduler(%v) stop with error:%v", errInfo.desc, errInfo.err)
		log.Errorf(err.Error())
		return err
	}
}

func (scd *Scheduler) Stop() {
	for _, app := range scd.apps {
		app.stop()
	}
}

// WithScheduler 添加调度器
func (scd *Scheduler) withAppDescInfo(adi *ApplicationDescInfo) *Scheduler {
	scd.adis = append(scd.adis, adi)
	return scd
}

// initialize 初始化app
func (scd *Scheduler) initialize() error {
	var schedulerBootFlags []interface{}
	for _, v := range scd.adis {
		curApp := newApp(v.name, v.options...)
		scd.apps = append(scd.apps, curApp)
		if curApp.bootFlag == nil {
			continue
		}
		schedulerBootFlags = append(schedulerBootFlags, curApp.bootFlag)
	}

	// 解析启动参数
	flags.ParseWithStructPointers(append([]interface{}{scd.globalBootFlag}, schedulerBootFlags...)...)

	// 解析配置文件
	if scd.globalBootConfigFileContent != nil && scd.globalBootFlag.BootConfigFile != "" {

		content, err := ioutil.ReadFile(scd.globalBootFlag.BootConfigFile)
		if err != nil {
			newErr := fmt.Errorf("load boot config file %v error:%v", scd.globalBootFlag.BootConfigFile, err)
			return newErr
		}

		err = scd.globalBootConfigParser(content, scd.globalBootConfigFileContent)
		if err != nil {
			newErr := fmt.Errorf("load boot config file %v content %v ok, but parse content error:%v",
				scd.globalBootFlag.BootConfigFile, string(content), err)
			return newErr
		}
	}

	// 检查一下启动参数
	if scd.globalBootFlag.ServiceName != "" {
		// 可能为pod name，解析-前面的deployment名字
		scd.globalBootFlag.ServiceName = strings.Split(scd.globalBootFlag.ServiceName, "-")[0]
	} else {
		// 否则就用二进制程序名字
		scd.globalBootFlag.ServiceName = filepath.Base(os.Args[0])
	}
	if scd.globalBootFlag.GlobalID == "" {
		newGlobalID, newServiceName := utils.GetGlobalIDName(scd.globalBootFlag.ServiceName)
		scd.globalBootFlag.GlobalID = newGlobalID
		scd.globalBootFlag.ServiceName = newServiceName
	}

	// 初始化日志系统
	var logHandlers []io.Writer
	if scd.globalBootFlag.LogDirPath == "" {
		// 没有指定日志输出目录，默认输出到控制台
		logHandlers = append(logHandlers, os.Stdout)
	} else {
		// 指定日志输出目录
		logHandler, err := handler.NewRotatingDayMaxFileHandler(scd.globalBootFlag.LogDirPath, scd.globalBootFlag.ServiceName, 1<<30, 10)
		if err != nil {
			newErr := fmt.Errorf("new log file handler with path [%v] name[%v] error:%v",
				scd.globalBootFlag.LogDirPath, scd.globalBootFlag.ServiceName, err)
			return newErr
		}
		logHandlers = append(logHandlers, logHandler)

		// 也指定输出到控制台
		if scd.globalBootFlag.LogStdout {
			logHandlers = append(logHandlers, os.Stdout)
		}
	}

	// 获取日志等级默认日志等级0，对应zerolog是debug
	logLevel := scd.defaultLogLevel
	if scd.globalBootFlag.LogLevel != "" {
		logLevel = log.LogLevelStr2Enum[scd.globalBootFlag.LogLevel]
	}
	// 创建logger
	log.NewGlobalLogger(logHandlers, logLevel, func(l zerolog.Logger) zerolog.Logger {
		return l.With().Str("service", scd.globalBootFlag.ServiceName).Str("node_id", scd.globalBootFlag.GlobalID).Logger()
	})

	// 初始化prometheus metrics、go pprof
	scd.server = prom.NewEngine(":"+scd.globalBootFlag.TracePort, true)

	// 初始化holmes dump
	holmesPath := scd.globalBootFlag.LogDirPath
	if holmesPath != "" {
		if holmesPath[len(holmesPath)-1] != '/' {
			holmesPath += "/"
		}
	}
	holmes.StartTraceAndDump(holmesPath + "holmes/" + scd.globalBootFlag.ServiceName)

	return nil
}

func (scd *Scheduler) initApps() error {
	for i, app := range scd.apps {
		if scd.adis[i].initFunc != nil {
			err := scd.adis[i].initFunc(scd.globalBootFlag, scd.globalBootConfigFileContent, app)
			if err != nil {
				return fmt.Errorf("application[%v] init return error[%v]", app.Name, err)
			} else {
				log.Noticef("application[%v] initialize ok", app.Name)
			}
		}
	}

	return nil
}

func (scd *Scheduler) applyOptions(options ...SchedulerOption) *Scheduler {
	for _, option := range options {
		option.Apply(scd)
	}
	return scd
}
