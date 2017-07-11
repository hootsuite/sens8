package handler

import (
	"testing"

	"github.com/upfluence/base/monitoring"
	"github.com/upfluence/goutils/monitoring/metric"
	"github.com/upfluence/goutils/monitoring/metric/mock"
)

func TestHandlerWithMetric(t *testing.T) {
	metrics := map[monitoring.MetricID]metric.Metric{
		"m1": mock.NewMockMetric(1.0, []string{""}),
		"m2": mock.NewMockMetric(2.0, []string{"foo", "bar"}),
		"m3": mock.NewMockMetric(3.0, []string{""}),
	}

	query := []string{"m1", "m2"}

	handler := NewMonitoringHandler("hey", metrics)

	res, err := handler.Collect(query)

	if err != nil {
		t.Errorf("Expected a success, got [%s]", err)
	}

	if l := len(res); l != 3 {
		t.Errorf("Expected 2 results, got [%d]", l)
	}

	if r, ok := res["hey.m1"]; ok && r != 1.0 {
		t.Errorf("Expected 1.0 for m1 value, got [%f]", r)
	} else if !ok {
		t.Error("Expected to find result for metric m1")
	}

	if r, ok := res["hey.m2.foo"]; ok && r != 2.0 {
		t.Errorf("Expected 2.0 for m2 value, got [%f]", r)
	} else if !ok {
		t.Error("Expected to find result for metric m2")
	}

	if r, ok := res["hey.m2.bar"]; ok && r != 2.0 {
		t.Errorf("Expected 2.0 for m2 value, got [%f]", r)
	} else if !ok {
		t.Error("Expected to find result for metric m2")
	}
}

func TestHandlerWithUnknownMetric(t *testing.T) {
	metrics := map[monitoring.MetricID]metric.Metric{
		"m1": mock.NewMockMetric(1.0, []string{""}),
		"m2": mock.NewMockMetric(2.0, []string{""}),
		"m3": mock.NewMockMetric(3.0, []string{""}),
	}

	query := []string{"m1", "m4"}

	handler := NewMonitoringHandler("hey", metrics)

	res, err := handler.Collect(query)

	if err == nil {
		t.Error("Expected an error here")
	}

	if res != nil {
		t.Error("We don't expect a result here")
	}
}
