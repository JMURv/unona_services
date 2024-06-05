package prometheus

import (
	"context"
	"fmt"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Metric struct {
	reg      *prometheus.Registry
	metrics  *grpcprom.ServerMetrics
	exemplar func(ctx context.Context) prometheus.Labels
}

func New() *Metric {
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		srvMetrics,
		PanicsTotal,
		RequestMetrics,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metric{
		reg:     reg,
		metrics: srvMetrics,
		exemplar: func(ctx context.Context) prometheus.Labels {
			return prometheus.Labels{"traceID": strconv.Itoa(1)}
		},
	}
}

func (m *Metric) Start(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		m.reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	log.Printf("starting http server for prometheus on port:%d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (m *Metric) ConfigureServerGRPC() *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			m.metrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(m.exemplar)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(PanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			m.metrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(m.exemplar)),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(PanicRecoveryHandler)),
		),
	)
}

func (m *Metric) Close() {}

var PanicsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "users",
	Name:      "panics_recovered_total",
	Help:      "Total number of gRPC requests recovered from internal panic.",
})

var PanicRecoveryHandler = func(p any) (err error) {
	PanicsTotal.Inc()
	return status.Errorf(codes.Internal, "%s", p)
}

var RequestMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  "users",
	Name:       "request_metrics",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
}, []string{"status", "endpoint"})

func ObserveRequest(d time.Duration, status int, endpoint string) {
	RequestMetrics.WithLabelValues(strconv.Itoa(status), endpoint).Observe(d.Seconds())
}
