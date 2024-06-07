package jaeger

import (
	cfg "github.com/JMURv/unona/services/pkg/config"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"io"
	"log"
)

type Tracing struct {
	closer io.Closer
}

func New(serviceName string, conf *cfg.JaegerConfig) *Tracing {
	tracerCfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  conf.Sampler.Type,
			Param: float64(conf.Sampler.Param),
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           conf.Reporter.LogSpans,
			LocalAgentHostPort: conf.Reporter.LocalAgentHostPort,
		},
	}

	tracer, closer, err := tracerCfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatalf("Error initializing Jaeger tracer: %s", err.Error())
	}

	opentracing.SetGlobalTracer(tracer)
	return &Tracing{
		closer: closer,
	}
}

func (t Tracing) Close() error {
	if err := t.closer.Close(); err != nil {
		return err
	}
	return nil
}
