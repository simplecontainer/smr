package metrics

import "github.com/prometheus/client_golang/prometheus"

type Counter struct {
	metric *prometheus.CounterVec
}

func (c *Counter) Register() {
	prometheus.MustRegister(c.metric)
}

func (c *Counter) Increment(labels ...string) {
	c.metric.WithLabelValues(labels...).Inc()
}

type Gauge struct {
	metric *prometheus.GaugeVec
}

func (g *Gauge) Register() {
	prometheus.MustRegister(g.metric)
}

func (g *Gauge) Set(value float64, labels ...string) {
	g.metric.WithLabelValues(labels...).Set(value)
}

type Histogram struct {
	metric *prometheus.HistogramVec
}

func (h *Histogram) Register() {
	prometheus.MustRegister(h.metric)
}

func (h *Histogram) Observe(value float64, labels ...string) {
	h.metric.WithLabelValues(labels...).Observe(value)
}
