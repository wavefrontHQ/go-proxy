package agent

import (
	"github.com/wavefronthq/go-proxy/api"
	"time"
	"log"
)

type WavefrontAgent interface {
	InitAgent()
}

type DefaultAgent struct {
	AgentID    string
	ApiService api.WavefrontAPI
}

func (agent *DefaultAgent) InitAgent() {
	// fetch configuration once per minute
	checkinTicker := time.NewTicker(time.Minute * time.Duration(1))
	go agent.checkin(checkinTicker)
}

func (agent *DefaultAgent) checkin(ticker * time.Ticker) {
	for range ticker.C {
		log.Println("Fetching configuration")
		agent.doFetchConfig()
	}
}

func (agent *DefaultAgent) doFetchConfig() {
	//TODO: fetch config from server periodically (figure out parameters) and update forwarder
	// parameters should include the latest agent metrics
	currentTime := getCurrentTime()
	agentConfig, err := agent.ApiService.GetConfig(currentTime, 0, 0, 0)
	if err != nil {
		log.Println("Error fetching config", err)
		return
	}
	log.Println("AgentConfig", *agentConfig)
	agent.ApiService.AgentConfigProcessed()
}

func getCurrentTime() int64 {
	return time.Now().UnixNano()/1000000000
}
