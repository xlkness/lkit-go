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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var CommonBootFlag = &CommBootFlag{}

type CommBootFlag struct {
	GlobalID       string `name:"global_id" desc:"全局唯一id，为空会给随机字符串" default:""`
	ServiceName    string `name:"service_name" desc:"当前进程服务名，为空会用当前可执行文件名" default:""`
	BootConfigFile string `name:"boot_config_file" desc:"起服配置文件路径，例如：/dir/boot_config.yaml" default:""`
	TracePort      string `name:"trace_port" desc:"监控端口，包含prometheus、go pprof等" default:"7788"`
	LogDirPath     string `name:"log_dir" desc:"程序日志输出目录" default:"log"`
	LogStdout      bool   `name:"log_stdout" desc:"程序日志是否也输出到控制台" default:"false"`
	LogLevel       string `name:"log_level" desc:"trace|debug|info|notice|warn|error|criti|fatal|panic" default:""`
}

type Scheduler struct {
	globalBootFlag              *CommBootFlag                          // app全局启动参数
	globalBootConfigFileContent interface{}                            // app全局配置文件内容结构体指针，为空没有配置文件解析
	globalBootConfigParser      func(in []byte, out interface{}) error // 配置文件解析函数，默认yaml
	defaultLogLevel             log.LogLevel                           // 默认debug日志等级，优先用globalBootFlag指定的日志等级
	apps                        []pair                                 // 可绑定多个调度器
	server                      *engine.Engine                         // app全局的web服务，当前暂时一个，为prometheus、pprof共用
}

// NewScheduler
func NewScheduler(appOptions ...SchedulerOption) *Scheduler {
	scd := new(Scheduler)
	scd.globalBootFlag = CommonBootFlag
	scd.applyOptions(appOptions...)
	return scd
}

// WithScheduler 添加调度器
func (scd *Scheduler) WithApp(desc string, app *App) *Scheduler {
	scd.apps = append(scd.apps, pair{desc, app})
	return scd
}

func (scd *Scheduler) Run() error {
	// 初始化app
	err := scd.initialize()
	if err != nil {
		return err
	}

	// 启动调度器
	type waitInfo struct {
		desc string
		scd  *App
		err  error
	}
	waitChan := make(chan waitInfo, 1)
	for _, scd := range scd.apps {
		go func(desc string, scheduler *App) {
			err := scheduler.run()
			if err != nil {
				// 返回调度器的报错
				waitChan <- waitInfo{desc, scheduler, err}
			}
		}(scd.desc, scd.item.(*App))
	}

	// 运行trace server
	go func() {
		log.Noticef("trace server listen on:%v", scd.server.Addr)
		err := scd.server.Run()
		if err != nil {
			waitChan <- waitInfo{"trace server", nil, err}
		}
	}()

	watchSignChan := libsyscal.WatchSignal1()

	defer scd.Stop()

	select {
	case signal := <-watchSignChan:
		log.Noticef("application receive signal(%v), will graceful stop", signal)
		return nil
	case errInfo := <-waitChan:
		err := fmt.Errorf("application receive scheduler(%v) stop with error:%v", errInfo.desc, errInfo.err)
		log.Errorf(err.Error())
		return err
	}
}

func (scd *Scheduler) Stop() {
	for _, app := range scd.apps {
		app.item.(*App).stop()
	}
}

// initialize 初始化app
func (scd *Scheduler) initialize() error {
	var schedulerBootFlags []interface{}
	for _, v := range scd.apps {
		schduler := v.item.(*App)
		if schduler.bootFlag == nil {
			continue
		}
		schedulerBootFlags = append(schedulerBootFlags, schduler.bootFlag)
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
	logHandler, err := handler.NewRotatingDayMaxFileHandler(scd.globalBootFlag.LogDirPath, scd.globalBootFlag.ServiceName, 1<<30, 10)
	if err != nil {
		newErr := fmt.Errorf("new log file handler with path [%v] name[%v] error:%v",
			scd.globalBootFlag.LogDirPath, scd.globalBootFlag.ServiceName, err)
		return newErr
	}
	// 获取日志等级默认日志等级0，对应zerolog是debug
	logLevel := scd.defaultLogLevel
	if scd.globalBootFlag.LogLevel != "" {
		logLevel = log.LogLevelStr2Enum[scd.globalBootFlag.LogLevel]
	}
	// 创建logger
	log.NewGlobalLogger(logHandler, logLevel, func(l zerolog.Logger) zerolog.Logger {
		return l.With().Str("service", scd.globalBootFlag.ServiceName).Str("node_id", scd.globalBootFlag.GlobalID).Logger()
	}, scd.globalBootFlag.LogStdout)

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

func (scd *Scheduler) applyOptions(options ...SchedulerOption) *Scheduler {
	for _, option := range options {
		option.Apply(scd)
	}
	return scd
}
