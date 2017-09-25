package agent

import (
	"encoding/json"
	"fmt"

	"github.com/rcrowley/go-metrics"
)

func buildAgentMetrics() ([]byte, error) {
	var stats map[string]interface{} = make(map[string]interface{})
	metrics.DefaultRegistry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			stats[name] = metric.Count()
		case metrics.Gauge:
			stats[name] = metric.Value()
		case metrics.GaugeFloat64:
			stats[name] = metric.Value()
		case metrics.Timer:
			timer := metric.Snapshot()
			addHisto(stats, name, timer.Min(), timer.Max(), timer.Mean(),
				timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}))
			addRate(stats, name, timer.Count(), timer.Rate1(), timer.RateMean())
		case metrics.Histogram:
			histo := metric.Snapshot()
			addHisto(stats, name, histo.Min(), histo.Max(), histo.Mean(),
				histo.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}))
		case metrics.Meter:
			meter := metric.Snapshot()
			addRate(stats, name, meter.Count(), meter.Rate1(), meter.RateMean())
		}
	})
	return json.Marshal(stats)
}

func addHisto(stats map[string]interface{}, name string, min, max int64, mean float64, percentiles []float64) {
	// convert from nanoseconds to milliseconds
	stats[combine(name, "duration.min")] = float64(min) / 1e6
	stats[combine(name, "duration.max")] = float64(max) / 1e6
	stats[combine(name, "duration.mean")] = mean / 1e6
	stats[combine(name, "duration.median")] = percentiles[0] / 1e6
	stats[combine(name, "duration.p75")] = percentiles[1] / 1e6
	stats[combine(name, "duration.p95")] = percentiles[2] / 1e6
	stats[combine(name, "duration.p99")] = percentiles[3] / 1e6
	stats[combine(name, "duration.p999")] = percentiles[4] / 1e6
}

func addRate(stats map[string]interface{}, name string, count int64, m1, mean float64) {
	stats[combine(name, "rate.count")] = count
	stats[combine(name, "rate.m1")] = m1
	stats[combine(name, "rate.mean")] = mean
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}
