package lkit_go

import (
	"fmt"
	"testing"
)

type TestBootFileContent struct {
	Region string `yaml:"region"`
	Server string `yaml:"server"`
	Dsn    string `yaml:"dsn"`
}

type TestApp1BootFlag struct {
	Port string `env:"port" desc:"listen port" default:"8888"`
}

type TestApp2BootFlag struct {
	ID string `env:"id" desc:"unique id" default:"1234"`
}

func getApp1InitDescInfo() *ApplicationDescInfo {
	appBootFlag := &TestApp1BootFlag{}
	option := WithAppBootFlag(appBootFlag)
	appDescInfo := NewApplicationDescInfo("app1", func(globalBootFlag *CommBootFlag, globalBootFile interface{}, app *Application) error {
		if appBootFlag.Port != "" {
			// todo 启动监听服务器
		}
		app.WithInitializeTask("检查某些逻辑", func() error {
			fmt.Printf("app1 boot flag[%v] check...\n", appBootFlag.Port)
			return nil
		})
		return nil
	}, option)
	return appDescInfo
}

func getApp2InitDescInfo() *ApplicationDescInfo {
	appBootFlag := &TestApp2BootFlag{}
	option := WithAppBootFlag(appBootFlag)
	appDescInfo := NewApplicationDescInfo("app2", func(globalBootFlag *CommBootFlag, globalBootFile interface{}, app *Application) error {
		if appBootFlag.ID != "" {
			// todo some logic
		}
		app.WithInitializeTask("检查某些逻辑", func() error {
			fmt.Printf("app2 boot flag[%v] check...\n", appBootFlag.ID)
			return nil
		})
		return nil
	}, option)
	return appDescInfo
}

func TestApplication(t *testing.T) {
	// 获取app1的描述信息
	app1 := WithSchedulerCreateOneAppOption(getApp1InitDescInfo())
	// 获取app2的描述信息
	app2 := getApp2InitDescInfo()
	// 创建调度器，并且app1以构建参数传入
	scd := NewScheduler(WithSchedulerLogFileLevel(LogLevelTrace), WithSchedulerBootConfigFileContent(&TestBootFileContent{}), app1)
	// app2调用CreateApp来创建添加到调度器上
	scd.CreateApp(app2)
	// 启动调度器
	err := scd.Run()
	if err != nil {
		panic(err)
	}
}
