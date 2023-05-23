package new

var tplService = `
package service

import (
	"context"
	api_{{.AppName}} "{{.AppName}}/api"
	"github.com/xlkness/lkit-go/joyservice"
)

type Service struct {
}

func (svc *Service) Say(context.Context, *api_{{.AppName}}.{{.AppCamelName}}Req, *api_{{.AppName}}.{{.AppCamelName}}Res) error {
	return nil
}

func New() (*joyservice.ServicesManager, error) {
	svc := new(Service)
	svcMgr, err := api_hello.New{{.AppCamelName}}Handler(":8080", ":8080", []string{":2379"}, svc, true)
	if err != nil {
		return nil, err
	}
	return svcMgr, nil
}

`
