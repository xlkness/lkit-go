package rpc_tracer

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// func NewJaegerTracer(reporterServerAddr, serviceName string) (opentracing.Tracer, error) {
// 	cfg := &jaegerConfig.Configuration{
// 		Sampler: &jaegerConfig.SamplerConfig{
// 			Type:  jaeger.SamplerTypeConst, //固定采样
// 			Param: 1,                       //1=全采样、0=不采样
// 		},
//
// 		Reporter: &jaegerConfig.ReporterConfig{
// 			LogSpans:           true,
// 			LocalAgentHostPort: reporterServerAddr,
// 		},
//
// 		ServiceName: serviceName,
// 	}
//
// 	tracer, _, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
// 	if err != nil {
// 		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
// 	}
//
// 	opentracing.SetGlobalTracer(tracer)
//
// 	return tracer, nil
// }

func NewJaegerTracerProvider(service string, id string, jaegerUrl string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			// attribute.String("environment", env),
			attribute.String("ID", id),
		)),
	)
	return tp, nil
}
