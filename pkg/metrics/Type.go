package metrics

// MetricType is an enum-like type to represent different types of metrics.
type MetricType string

const (
	CounterType   MetricType = "counter"
	GaugeType     MetricType = "gauge"
	HistogramType MetricType = "histogram"
)

// Metric is an interface that all Prometheus metrics must implement.
type Metric interface {
	Register()
}
