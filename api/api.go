package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/wavefronthq/go-proxy/config"
)

var client = &http.Client{Timeout: time.Second * 10}
var pointError = errors.New("Invalid points")

type WavefrontAPI interface {
	GetConfig(currentMillis, bytesLeft, bytesPerMinute, currentQueueSize int64) (*config.AgentConfig, error)
	Checkin(currentMillis int64, localAgent, pushAgent, ephemeral bool, agentMetrics []byte) (*config.AgentConfig, error)
	PostData(workUnitId, format, pointLines string) (*http.Response, error)
	AgentError(details string)
	AgentConfigProcessed() error
}

type WavefrontAPIService struct {
	ServerURL string
	AgentID   string
	Hostname  string
	Token     string
	Version   string
}

func (service *WavefrontAPIService) GetConfig(currentMillis, bytesLeft, bytesPerMinute, currentQueueSize int64) (*config.AgentConfig, error) {
	log.Println("GetConfig")

	apiURL := service.ServerURL + GET_CONFIG_SUFFIX
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return &config.AgentConfig{}, err
	}

	q := req.URL.Query()
	q.Add(HOSTNAME, service.Hostname)
	q.Add(TOKEN, service.Token)
	q.Add(VERSION, service.Version)
	q.Add(CURRENT_MILLIS, strconv.FormatInt(currentMillis, 10))
	q.Add(BYTES_LEFT, strconv.FormatInt(bytesLeft, 10))
	q.Add(BYTES_PER_MIN, strconv.FormatInt(bytesPerMinute, 10))
	q.Add(CURR_QUEUE_SIZE, strconv.FormatInt(currentQueueSize, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &config.AgentConfig{}, err
	}
	defer resp.Body.Close()

	config := &config.AgentConfig{}
	err = json.NewDecoder(resp.Body).Decode(config)
	return config, err
}

func (service *WavefrontAPIService) Checkin(currentMillis int64, localAgent, pushAgent, ephemeral bool, agentMetrics []byte) (*config.AgentConfig, error) {
	log.Println("Checkin")

	apiURL := service.ServerURL + CHECKIN_SUFFIX
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(agentMetrics))
	req.Header.Set(CONTENT_TYPE, APPLICATION_JSON)
	if err != nil {
		return &config.AgentConfig{}, err
	}

	q := req.URL.Query()
	q.Add(HOSTNAME, service.Hostname)
	q.Add(TOKEN, service.Token)
	q.Add(VERSION, service.Version)
	q.Add(CURRENT_MILLIS, strconv.FormatInt(currentMillis, 10))
	q.Add(LOCAL, strconv.FormatBool(localAgent))
	q.Add(PUSH, strconv.FormatBool(pushAgent))
	q.Add(EPHEMERAL, strconv.FormatBool(ephemeral))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &config.AgentConfig{}, err
	}
	defer resp.Body.Close()

	config := &config.AgentConfig{}
	err = json.NewDecoder(resp.Body).Decode(config)
	return config, err
}

func (service *WavefrontAPIService) PostData(workUnitId, format, pointLines string) (*http.Response, error) {
	if pointLines == "" {
		return &http.Response{}, pointError
	}

	apiURL := service.ServerURL + POST_DATA_SUFFIX
	apiURL = fmt.Sprintf(apiURL, service.AgentID, workUnitId, format)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(pointLines))
	req.Header.Set(CONTENT_TYPE, TEXT_PLAIN)
	if err != nil {
		return &http.Response{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	return resp, nil
}

func (service *WavefrontAPIService) AgentError(details string) {
	log.Println("AgentError")
}

func (service *WavefrontAPIService) AgentConfigProcessed() error {
	log.Println("AgentConfigProcessed")

	apiURL := service.ServerURL + CONFIG_PROCESSED_SUFFIX
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	return err
}
