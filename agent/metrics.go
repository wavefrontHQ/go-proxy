package agent

import (
	"encoding/json"
	"github.com/rcrowley/go-metrics"
	"log"
)

func buildAgentMetrics() ([]byte, error) {
	var agentMetrics map[string]interface{} = make(map[string]interface{})
	metrics.DefaultRegistry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			log.Println("Counter", name, metric.Count())
			agentMetrics[name] = metric.Count()
		}
		//TODO: expand to support other types (timer, gauge etc)
	})
	return json.Marshal(agentMetrics)
}
