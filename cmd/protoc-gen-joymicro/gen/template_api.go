package gen

var tempApiText = `
package {{ .Package }}

import (
	"context"
	"sync"
	"time"
	"reflect"
	"joynova.com/joynova/joymicro/joyclient"
	"joynova.com/joynova/joymicro/log"
	"github.com/smallnest/rpcx/client"
)

var _ = joyclient.Service{}

{{ $isEnableCHash := .IsEnableInvokeConsistentHash }}
{{ $isEnablePeer := .IsEnableSpecInvokePeer }}
{{ $hasKeyInvoke := or $isEnableCHash $isEnablePeer }}

var {{ .ServiceName_fooBar }}ServiceInstance {{ .ServiceInterfaceName }}

{{ $serviceReceiver := printf "%s%s" .ServiceName_fooBar "Service" }}

// LazyInit{{ .ServiceName_FooBar }}Service 懒汉模式初始化服务调用实例，只有在真正发生调用时才初始化
func LazyInit{{ .ServiceName_FooBar }}Service(etcdAddrs []string, timeout time.Duration, isPermanent, isLocal bool) {
	lazyInit{{ .ServiceName_FooBar }}ServiceFun = func() {
	c := New{{ .ServiceName_FooBar }}ServiceInstance(etcdAddrs, timeout, isPermanent, isLocal)
	{{- if $hasKeyInvoke }}
	if !isLocal {
		{{ if $isEnableCHash -}} 
		// 打开一致性hash调用，后续方法需要加入hash的key，相同key可以打到同一节点调用
		c.(*{{ $serviceReceiver }}).c.SetSelector(joyclient.NewConsistentHashSelector())
		{{ else if $isEnablePeer -}}
		// 打开点对点调用，后续方法需要加入key，根据key来匹配相同的节点调用
		c.(*{{ $serviceReceiver }}).c.SetSelector(new(joyclient.PeerSelector))
		{{- end -}}
	}
	{{ end }}
	{{ .ServiceName_fooBar }}ServiceInstance = c
	}
}

{{ $serviceFooBar := .ServiceName_FooBar }}
{{ $singletonInstance := printf "%s%s%s" "Get" .ServiceName_FooBar "ServiceInstance()" }}
{{ range $idx, $service := .Services }}
{{ range $idx1, $method := $service.Methods }}
{{ if $hasKeyInvoke }}
// {{ $method.Name }} key为指定
func {{ $method.Name }}(ctx context.Context, key string, in *{{ $method.InputType }}) (*{{ $method.OutputType }}, error) {
	ctx = context.WithValue(ctx, "select_key", key)
	instance := {{ $singletonInstance }}
	res, err := instance.{{ $method.Name }}(ctx, in)
	handle{{ $serviceFooBar }}CallError("{{ $method.Name }}", in, res, err)
	return res, err
}
{{ else }}
func {{ $method.Name }}(ctx context.Context, in *{{ $method.InputType }}) (*{{ $method.OutputType }}, error) {
	instance := {{ $singletonInstance }}
	res, err := instance.{{ $method.Name }}(ctx, in)
	handle{{ $serviceFooBar }}CallError("{{ $method.Name }}", in, res, err)
	return res, err
}
{{ end }}
{{ end }}
{{ end }}

// 真正使用时初始化
var {{ .ServiceName_fooBar }}ServiceInitOnce = new(sync.Once)
var lazyInit{{ .ServiceName_FooBar }}ServiceFun func()
func Get{{ .ServiceName_FooBar}}ServiceInstance() {{ .ServiceInterfaceName }} {
	if lazyInit{{ .ServiceName_FooBar }}ServiceFun != nil {
		{{ .ServiceName_fooBar }}ServiceInitOnce.Do(lazyInit{{ .ServiceName_FooBar }}ServiceFun)
	} else {
		return nil
	}
	return {{ .ServiceName_fooBar }}ServiceInstance
}

// handleError 以下逻辑用于rpc遇到call底层报错（例如服务器节点找不到），
// 调用返回了error，但是res里的errCode字段就没有赋值
func handle{{ .ServiceName_FooBar }}CallError(method string, req, res interface{}, err error) {
	if err == nil {
		return
	}
	if err.Error() == client.ErrXClientNoServer.Error() {
		log.Errorf("rpc call method(%v) with request(%+v) not found any server", method, req)
	}
	vo := reflect.ValueOf(res).Elem()
	to := reflect.TypeOf(res).Elem()
	if to.NumField() > 0 {
		// 字段数大于0，且第一个字段是整型（err_code会变为整型）
		if to.Field(0).Type.Kind() == reflect.Int32 {
			if vo.Field(0).Int() == 0 && vo.Field(0).CanSet() {
				vo.Field(0).SetInt(13579)
			}
		}
	}
}
`
