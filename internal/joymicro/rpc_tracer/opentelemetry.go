package rpc_tracer

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var TracerName = "rpcx" // 可以用来区分环境

var (
	onceJaeger           = new(sync.Once)
	onceProm             = new(sync.Once)
	jaegerTracerProvider *tracesdk.TracerProvider
)

// EnableJaegerTrace 打开jaeger调用链追踪，传进服务名、服务id、jaeger服务器地址
func EnableJaegerTrace(svc string, id1 string, jaegerUrl1 string) {
	onceJaeger.Do(func() {
		initFun(svc, id1, jaegerUrl1)
	})
}

func GetJaegerTracerProvider() *tracesdk.TracerProvider {
	return jaegerTracerProvider
}

func initFun(svc string, id string, jaegerUrl string) {
	tp, err := NewJaegerTracerProvider(svc, id, jaegerUrl)
	if err != nil {
		panic(fmt.Errorf("NewJaegerTracerProvider error:%v", err))
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tp)

	jaegerTracerProvider = tp
}
