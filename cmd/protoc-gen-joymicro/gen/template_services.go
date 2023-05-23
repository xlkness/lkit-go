package gen

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
func New{{ .ServiceName_FooBar }}ServiceInstance(etcdAddrs []string, timeout time.Duration, isPermanent, isLocal bool) {{ .ServiceInterfaceName }} {
if !isLocal {
	c := joyclient.New({{ $callServiceName }}, etcdAddrs, timeout, isPermanent)
	return &{{ $serviceReceiver }} {
		c: c,
	} 
}

// 本地函数调用模式，用于调试
return &{{ .ServiceName_fooBar }}ServiceLocal{}
}

// Set{{ .ServiceName_FooBar }}ServiceSelector 设置调用插件，可以用来监听服务节点变化、按需选择某个节点调用、自定义负载均衡算法等
func Set{{ .ServiceName_FooBar }}ServiceSelector(c {{ .ServiceInterfaceName }}, selector joyclient.Selector) {
c1, ok := c.(*{{ .ServiceName_fooBar }}Service)
if ok {
	c1.c.SetSelector(selector)
}
}

// {{ $serviceReceiver }} 调用服务的远程调用具体实现
type {{ $serviceReceiver }} struct {
	c *joyclient.Service
}

{{ range $idx, $service := .Services }}
{{ range $idx1, $method := $service.Methods }}
func (c *{{ $serviceReceiver }}) {{ $method.Name }}(ctx context.Context, in *{{ $method.InputType }}) (*{{ $method.OutputType }}, error) {
	var err error

	var out *{{ $method.OutputType }}
	// one way模式，消息只发到对端，不关心响应
	if v := ctx.Value("one_way"); v == nil || !v.(bool) {
		out = new({{ $method.OutputType }})
	}

	if v := ctx.Value("to_all_nodes"); v != nil && v.(bool) {
		err = c.c.CallAll(ctx, "{{ $method.Name }}", in, out)
		return out, err
	}

	err = c.c.Call(ctx, "{{ $method.Name }}", in, out)
	return out, err
}
{{ end }}
{{ end }}

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


// Register{{ .ServiceName_FooBar }}Handler 手工给服务注册handler，但必须在s.Run之前调用，metadata是自定义的服务描述信息，会传递给服务调用客户端
func Register{{ .ServiceName_FooBar }}Handler(s *joyservice.ServicesManager, handler {{ .HandlerInterfaceName }}, metadata map[string]string) error {
// 如果是本地调试函数调用，设置全局handler
Set{{ .ServiceName_FooBar }}HandlerLocal(handler)
return s.RegisterOneService({{ $callServiceName }}, handler, metadata)
}

{{ $isEnablePeer := .IsEnableSpecInvokePeer }}
// New{{ .ServiceName_FooBar }}Handler 创建并注册、运行一个服务
{{ if $isEnablePeer }}
func New{{ .ServiceName_FooBar }}Handler(nodeKey, listenAddr, exposeAddr string, etcdAddrs []string, handler {{ .HandlerInterfaceName }}, isLocal bool) (*joyservice.ServicesManager, error) {
if !isLocal {
s, err := joyservice.NewWithKey(nodeKey, listenAddr, exposeAddr, etcdAddrs)
{{ else }}
func New{{ .ServiceName_FooBar }}Handler(listenAddr, exposeAddr string, etcdAddrs []string, handler {{ .HandlerInterfaceName }}, isLocal bool) (*joyservice.ServicesManager, error) {
if !isLocal {
s, err := joyservice.New(listenAddr, exposeAddr, etcdAddrs)
{{ end }}
	if err != nil {
		return nil, err
	}

	err = Register{{ .ServiceName_FooBar }}Handler(s, handler, nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// 如果是本地调试函数调用，设置全局handler
Register{{ .ServiceName_FooBar }}Handler(nil, handler, nil)
return nil, nil
}


// ============================================调试模式变量============================================

var {{ .ServiceName_fooBar }}HandlerLocal {{ .HandlerInterfaceName }}

func Set{{ .ServiceName_FooBar }}HandlerLocal(handler {{ .HandlerInterfaceName }}) {
{{ .ServiceName_fooBar }}HandlerLocal = handler
}

// {{ .ServiceName_fooBar }}ServiceLocal 本地函数调用，用于调试
type {{ .ServiceName_fooBar }}ServiceLocal struct {}

{{ $serviceLocalReceiver := printf "%s%s" .ServiceName_fooBar "ServiceLocal" }}
{{ $handlerLocalReceiver := printf "%s%s" .ServiceName_fooBar "HandlerLocal" }}

{{ range $idx, $service := .AllServices }}
{{ range $idx1, $method := $service.Methods }}
func (c *{{ $serviceLocalReceiver }}) {{ $method.Name }}(ctx context.Context, in *{{ $method.InputType }}) (*{{ $method.OutputType }}, error) {
var out *{{ $method.OutputType }}
var err error
if v1 := ctx.Value("one_way"); v1 == nil || !v1.(bool) {
	out = new({{ $method.OutputType }})
}
err = {{ $handlerLocalReceiver }}.{{ $method.Name }}(ctx, in, out)
return out, err
}
{{ end }}
{{ end }}

`
