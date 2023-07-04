package gen_all_in_one

var tempServicesText = tempfileCommonText + `

{{ $isEnableCHash := .IsEnableInvokeConsistentHash }}
{{ $isEnablePeer := .IsEnableSpecInvokePeer }}

// {{ .ServiceInterfaceName }} 服务调用接口
type {{ .ServiceInterfaceName }} interface {
{{ if .MultiServices }}
{{- range $idx, $service := .AllServices }}
{{ $service.ServiceInterfaceName }}
{{- end }}
}
{{ range $idx, $service := .AllServices }}
type {{ $service.ServiceInterfaceName }} interface {
{{- range $idx1, $method := $service.Methods }}
{{ $method.Name }}(context.Context, *{{ $method.InputType }}) (*{{ $method.OutputType }}, error)
{{- end -}}
}
{{ end }}
{{ else }}
{{ range $idx, $service := .AllServices }}
{{- range $idx1, $method := $service.Methods }}
{{ $method.Name }}(context.Context, *{{ $method.InputType }}) (*{{ $method.OutputType }}, error)
{{- end -}}
{{ end }}
}
{{ end }}

{{ $callServiceName := printf "\"%s\"" .ServiceName_fooBar }}
{{ $serviceReceiver := printf "%s%s" .ServiceName_fooBar "Service" }}
// New{{ .ServiceName_FooBar }}Service 创建服务调用
func New{{ .ServiceName_FooBar }}ServiceInstance() {{ .ServiceInterfaceName }} {
// 本地函数调用模式，用于调试
return &{{ .ServiceName_fooBar }}ServiceLocal{}
}

// {{ .HandlerInterfaceName }} 服务节点handler接口定义
type {{ .HandlerInterfaceName }} interface {
{{ if .MultiServices }}
{{- range $idx, $service := .AllServices }}
{{ $service.HandlerInterfaceName }}
{{- end -}}
}

{{ range $idx, $service := .AllServices }}
type {{ $service.HandlerInterfaceName }} interface {
{{- range $idx1, $method := $service.Methods }}
{{ $method.Name }}(context.Context, *{{ $method.InputType }}, *{{ $method.OutputType }}) error
{{- end -}}
}
{{ end }}
{{ else }}
{{ range $idx, $service := .AllServices }}
{{- range $idx1, $method := $service.Methods }}
{{ $method.Name }}(context.Context, *{{ $method.InputType }}, *{{ $method.OutputType }}) error
{{- end -}}
}
{{ end }}
{{ end }}

{{ $isEnablePeer := .IsEnableSpecInvokePeer }}
// New{{ .ServiceName_FooBar }}Handler 创建并注册、运行一个服务
func New{{ .ServiceName_FooBar }}Handler(handler {{ .HandlerInterfaceName }}) error {
{{ .ServiceName_fooBar }}HandlerLocal = handler
return nil
}


// ============================================调试模式变量============================================

var {{ .ServiceName_fooBar }}HandlerLocal {{ .HandlerInterfaceName }}

// {{ .ServiceName_fooBar }}ServiceLocal 本地函数调用，用于调试
type {{ .ServiceName_fooBar }}ServiceLocal struct {}

{{ $serviceLocalReceiver := printf "%s%s" .ServiceName_fooBar "ServiceLocal" }}
{{ $handlerLocalReceiver := printf "%s%s" .ServiceName_fooBar "HandlerLocal" }}

{{ range $idx, $service := .AllServices }}
{{ range $idx1, $method := $service.Methods }}
func (c *{{ $serviceLocalReceiver }}) {{ $method.Name }}(ctx context.Context, in *{{ $method.InputType }}) (*{{ $method.OutputType }}, error) {
var out *{{ $method.OutputType }} = new({{ $method.OutputType }})
var err error
err = {{ $handlerLocalReceiver }}.{{ $method.Name }}(ctx, in, out)
return out, err
}
{{ end }}
{{ end }}

`
