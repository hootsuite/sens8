package handler

import (
	"strings"

	"github.com/upfluence/base/monitoring"
	"github.com/upfluence/goutils/monitoring/metric"
)

type MonitoringHandler struct {
	prefix  string
	metrics map[monitoring.MetricID]metric.Metric
}

func NewMonitoringHandler(
	prefix string,
	metrics map[monitoring.MetricID]metric.Metric,
) *MonitoringHandler {
	return &MonitoringHandler{prefix, metrics}
}

// Can't understand why thrift inteface declares a []string instead of
// []monitoring.MetricsID
func (m *MonitoringHandler) Collect(metrics []string) (
	monitoring.Metrics,
	error,
) {
	promises := make(map[monitoring.MetricID]chan []metric.Point)
	results := monitoring.Metrics{}

	for _, id := range metrics {
		metricId := monitoring.MetricID(id)

		if met, ok := m.metrics[metricId]; ok {
			result := make(chan []metric.Point)
			promises[metricId] = result

			go func() {
				result <- met.Collect()
			}()
		} else {
			return nil, &monitoring.UnknownMetric{metricId}
		}
	}

	for id, promise := range promises {
		for _, point := range <-promise {
			splittedName := []string{m.prefix, string(id)}
			if v := point.Suffix; v != "" {
				splittedName = append(splittedName, v)
			}

			metricName := monitoring.MetricID(strings.Join(splittedName, "."))
			results[metricName] = point.Value
		}
	}

	return results, nil
}
