package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewCounter(name string, help string, labels []string) *Counter {
	counter := &Counter{
		metric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
	}
	counter.Register()
	return counter
}

func NewGauge(name string, help string, labels []string) *Gauge {
	gauge := &Gauge{
		metric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
	}

	gauge.Register()
	return gauge
}

func NewHistogram(name string, help string, labels []string) *Histogram {
	histogram := &Histogram{
		metric: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    name,
				Help:    help,
				Buckets: prometheus.DefBuckets,
			},
			labels,
		),
	}

	histogram.Register()
	return histogram
}
