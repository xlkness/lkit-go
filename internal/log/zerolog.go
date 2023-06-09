package log

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogLevel = zerolog.Level

var (
	LogLevelTrace  = zerolog.TraceLevel // 追踪日志，调试完毕需要删除
	LogLevelDebug  = zerolog.DebugLevel // 调试日志，可以放在代码中，线上出问题调成这个级别
	LogLevelInfo   = zerolog.InfoLevel  // 正常关键逻辑记录信息，线上日常设置为这个级别
	LogLevelNotice = zerolog.Level(99)  // 系统关键节点时输出的留意日志
	LogLevelWarn   = zerolog.WarnLevel  // 警告，某些逻辑出现意向不到的情况，输出告警，例如配置表错误、rpc错误
	LogLevelError  = zerolog.ErrorLevel // 错误，服务器重要组件出现意向不到的情况，输出错误，例如数据库、redis错误
	LogLevelCriti  = zerolog.Level(100) // 危急，用于需要开发注意的信息，例如崩溃但不影响服务器运行的栈日志，一般接上sms、im告警
	LogLevelFatal  = zerolog.FatalLevel // 致命，核心组建出问题，无法运行，输出告警，并以1的错误码退出
	LogLevelPanic  = zerolog.PanicLevel // 崩溃，核心组建出问题，无法运行，崩溃退出
)

var LogLevelStr2Enum = map[string]LogLevel{
	"trace":  LogLevelTrace,
	"debug":  LogLevelDebug,
	"info":   LogLevelInfo,
	"notice": LogLevelNotice,
	"warn":   LogLevelWarn,
	"error":  LogLevelError,
	"criti":  LogLevelCriti,
	"fatal":  LogLevelFatal,
	"panic":  LogLevelPanic,
}

func init() {
	// 设置时间格式
	zerolog.TimeFieldFormat = "06/01/02 15:04:05.000"
	// 修改level字段key，防止跟调用方的key一样
	zerolog.LevelFieldName = "log_level"
	// 修改时间key
	zerolog.TimestampFieldName = "log_time"
	// 设置调用文件名路径深度
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		depth := 1
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				if depth == FileReversedDepth {
					file = file[i+1:]
					break
				}
				depth++
			}
		}
		return file + ":" + strconv.Itoa(line)
	}
	// 设置全局日志等级
	zerolog.SetGlobalLevel(LogLevelTrace)
}

// 文件路径保留深度
var FileReversedDepth = 3

func NewGlobalLogger(writers []io.Writer, level LogLevel, initFun func(logger zerolog.Logger) zerolog.Logger) {
	// 设置全局日志等级
	zerolog.SetGlobalLevel(level)
	var parentLogger zerolog.Logger

	if len(writers) == 0 {
		fmt.Fprintf(os.Stderr, "NewGlobalLogger but write is nil, default give os.Stdout\n")
		writers = append(writers, os.Stdout)
	}

	multi := zerolog.MultiLevelWriter(writers...)

	// 创建全局日志
	parentLogger = zerolog.New(multi).With().Logger()

	if initFun != nil {
		log.Logger = initFun(parentLogger)
	} else {
		log.Logger = parentLogger
	}

	log.Logger = log.Hook(new(PrefixHook))
}

func NewCustomLogger(writer Handler, initFun func(logger *zerolog.Logger) *zerolog.Logger) *zerolog.Logger {
	if writer == nil {
		fmt.Fprintf(os.Stderr, "NewGlobalLogger but write is nil, default give os.Stdout\n")
		writer = os.Stdout
	}
	parentLogger := zerolog.New(writer).With().Logger()
	initFun(&parentLogger)
	return &parentLogger
}

// GetSubLogger 获取全局logger的子logger，可以设置子logger的输出格式
func GetSubLogger() zerolog.Context {
	return log.Logger.With()
}

func GetLogLevel() LogLevel {
	return log.Logger.GetLevel()
}

func SetLogLevel(level LogLevel) {
	zerolog.SetGlobalLevel(level)
}

func Tracef(format string, v ...interface{}) {
	output(LogLevelTrace, format, v...)
}

func Debugf(format string, v ...interface{}) {
	output(LogLevelDebug, format, v...)
}

func Infof(format string, v ...interface{}) {
	output(LogLevelInfo, format, v...)
}

func Noticef(format string, v ...interface{}) {
	output(LogLevelNotice, format, v...)
}

func Warnf(format string, v ...interface{}) {
	output(LogLevelWarn, format, v...)
}

func Errorf(format string, v ...interface{}) {
	output(LogLevelError, format, v...)
}

func Critif(format string, v ...interface{}) {
	output(LogLevelCriti, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	output(LogLevelFatal, format, v...)
}

func output(level LogLevel, format string, v ...interface{}) {
	var e *zerolog.Event
	switch level {
	case LogLevelTrace:
		e = log.Trace()
	case LogLevelDebug:
		e = log.Debug()
	case LogLevelInfo:
		e = log.Info()
	case LogLevelNotice:
		e = log.WithLevel(zerolog.NoLevel)
		e.Str("log_level", "notice")
	case LogLevelWarn:
		e = log.Warn()
	case LogLevelError:
		e = log.Error()
	case LogLevelCriti:
		e = log.WithLevel(zerolog.NoLevel)
		e.Str("log_level", "criti")
	case LogLevelFatal:
		e = log.Fatal()
	case LogLevelPanic:
		e = log.Panic()
	default:
		return
	}

	e.Timestamp().Caller(2).Msgf(format, v...)
}

func Output(level LogLevel) *zerolog.Event {
	var e *zerolog.Event
	switch level {
	case LogLevelTrace:
		e = log.Trace()
	case LogLevelDebug:
		e = log.Debug()
	case LogLevelInfo:
		e = log.Info()
	case LogLevelNotice:
		e = log.WithLevel(zerolog.NoLevel)
		e.Str("log_level", "notice")
	case LogLevelWarn:
		e = log.Warn()
	case LogLevelError:
		e = log.Error()
	case LogLevelCriti:
		e = log.WithLevel(zerolog.NoLevel)
		e.Str("log_level", "criti")
	case LogLevelFatal:
		e = log.Fatal()
	case LogLevelPanic:
		e = log.Panic()
	default:
		return nil
	}

	return e
}
