package new

import (
	"bytes"
	"fmt"
	"github.com/xlkness/lkit-go/cmd/lkit/utils"
	"github.com/xlkness/lkit-go/internal/cli"
	globalUtils "github.com/xlkness/lkit-go/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type commandNewFlag struct {
	SkipProtoStub bool `name:"skip-proto-stub" desc:"跳过proto协议桩代码生成" default:"false"`
}

func CommandNew() *cli.Command {
	cmd := cli.NewCommand("new", "new <app-name>，在当前目录生成app工程结构",
		"注意：生成工程需要协议构建，务必先安装protoc-3.0、gogofaster-gen-proto插件、joymicro-gen-proto插件",
		"./lkit new [-h|-help]", true, &commandNewFlag{}, commandNew)
	return cmd
}

func commandNew(cmd *cli.Command) error {
	if cmd.LastOptionArg == "" {
		utils.OutputError("%s", cmd.Usage(fmt.Sprintf("请输入创建的app名字！")))
		os.Exit(1)
	}

	appName := cmd.LastOptionArg
	flag := cmd.Flag.(*commandNewFlag)

	genInfo := &generateInfo{
		AppName:      appName,
		AppCamelName: globalUtils.CamelCase(appName),
	}

	utils.OutputInfo("准备创建应用：%v", appName)

	rootPath, _ := filepath.Abs("./")
	rootPath += "/" + appName + "/"
	utils.OutputInfo("应用生成目录：%v", rootPath)

	// 创建api
	apiPath := rootPath + "api"
	utils.OutputInfo("创建api目录：%s", apiPath)
	err := os.MkdirAll(apiPath, 0777)
	if err != nil {
		utils.OutputError("创建目录%s错误:%v", apiPath, err)
		os.Exit(1)
	}

	// 生成api
	err = generateFile(tplApiProto, genInfo, apiPath+"/"+appName+".proto")
	if err != nil {
		utils.OutputError("生成api协议文件错误：%s", err)
		os.Exit(1)
	}
	// 生成api桩代码
	if !flag.SkipProtoStub {
		execGenCmdStr := "protoc " + "-I" + apiPath + " " + "--gogofaster_out=. --joymicro_out=." + " " + apiPath + "/*.proto" + ""
		execGenCmd := exec.Command("protoc", "-I./", "--gogofaster_out=.", "--joymicro_out=.", appName+".proto")
		execGenCmd.Dir = apiPath
		if output, err := execGenCmd.CombinedOutput(); err != nil {
			utils.OutputError("execute "+execGenCmdStr+" output:\n%s\n[ERROR]%v\n", output, err)
		} else {
			utils.OutputInfo("生成api协议桩代码成功：%s", apiPath+"/"+appName+".proto")
		}
	}

	// 创建service
	servicePath := rootPath + "service"
	utils.OutputInfo("创建rpc服务目录：%s", servicePath)
	err = os.MkdirAll(servicePath, 0777)
	if err != nil {
		utils.OutputError("创建目录%s错误:%v", servicePath, err)
		os.Exit(1)
	}

	// 生成service代码
	generateFile(tplService, genInfo, servicePath+"/service.go")

	// 创建main文件
	generateFile(tplMain, genInfo, rootPath+"/main.go")

	// 生成go.mod
	generateFile(tplGoMod, genInfo, rootPath+"go.mod")
	generateFile(tplGoSum, genInfo, rootPath+"go.sum")

	utils.OutputInfo("服务[%s]生成完毕，目录结构如下：", appName)
	globalUtils.Tree(rootPath)

	utils.OutputInfo("执行以下指令可以启动服务：")
	utils.OutputInfo("  * 拉取依赖：go get github.com/xlkness/lkit-go")

	return nil
}

type generateInfo struct {
	AppName      string
	AppCamelName string
}

func generateFile(tplFile string, genInfo *generateInfo, path string) error {
	tmpl, err := template.New("lkit").Parse(tplFile)
	if err != nil {
		return fmt.Errorf("new tmpl with text:%s error:%v", tplFile, err)
	}

	//header := fmt.Sprintf("// Code generated by %s. DO NOT EDIT.\n", "lkit")

	buf := bytes.NewBuffer([]byte(""))
	err = tmpl.Execute(buf, genInfo)
	if err != nil {
		return fmt.Errorf("execute tmpl error:%v", err)
	}

	err = os.WriteFile(path, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("write buf to file(%v) error:%v", path, err)
	}

	return nil
}
