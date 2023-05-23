package new

var tplMain = `
package main

import (
	lkit_go "github.com/xlkness/lkit-go"
	"{{.AppName}}/service"
)

func main() {
	svc, err := service.New()
	if err != nil {
		panic(err)
	}
	
	app := lkit_go.NewApp()
	app.WithService("{{.AppName}}", svc)

	scd := lkit_go.NewScheduler()
	scd.WithApp("{{.AppName}}", app)
	err = scd.Run()
	if err != nil {
		panic(err)
	}
}
`
