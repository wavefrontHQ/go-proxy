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

var (
	client     = &http.Client{Timeout: time.Second * 30}
	pointError = errors.New("Invalid points")
)

// API interface for the agent.
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

	apiURL := service.ServerURL + getConfigSuffix
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return &config.AgentConfig{}, err
	}

	q := req.URL.Query()
	q.Add(hostnameParam, service.Hostname)
	q.Add(tokenParam, service.Token)
	q.Add(versionParam, service.Version)
	q.Add(currentMillisParam, strconv.FormatInt(currentMillis, 10))
	q.Add(bytesLeftParam, strconv.FormatInt(bytesLeft, 10))
	q.Add(bytesPerMinParam, strconv.FormatInt(bytesPerMinute, 10))
	q.Add(currentQueueSizeParam, strconv.FormatInt(currentQueueSize, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &config.AgentConfig{}, err
	}
	defer resp.Body.Close()

	cfg := &config.AgentConfig{}
	err = json.NewDecoder(resp.Body).Decode(cfg)
	return cfg, err
}

func (service *WavefrontAPIService) Checkin(currentMillis int64, localAgent, pushAgent, ephemeral bool, agentMetrics []byte) (*config.AgentConfig, error) {
	log.Println("Checkin")

	apiURL := service.ServerURL + checkinSuffix
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(agentMetrics))
	req.Header.Set(contentType, applicationJSON)
	if err != nil {
		return &config.AgentConfig{}, err
	}

	q := req.URL.Query()
	q.Add(hostnameParam, service.Hostname)
	q.Add(tokenParam, service.Token)
	q.Add(versionParam, service.Version)
	q.Add(currentMillisParam, strconv.FormatInt(currentMillis, 10))
	q.Add(localParam, strconv.FormatBool(localAgent))
	q.Add(pushParam, strconv.FormatBool(pushAgent))
	q.Add(ephemeralParam, strconv.FormatBool(ephemeral))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &config.AgentConfig{}, err
	}
	defer resp.Body.Close()

	cfg := &config.AgentConfig{}
	err = json.NewDecoder(resp.Body).Decode(cfg)
	return cfg, err
}

func (service *WavefrontAPIService) PostData(workUnitId, format, pointLines string) (*http.Response, error) {
	if pointLines == "" {
		return &http.Response{}, pointError
	}

	apiURL := service.ServerURL + postDataSuffix
	apiURL = fmt.Sprintf(apiURL, service.AgentID, workUnitId, format)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(pointLines))
	req.Header.Set(contentType, textPlain)
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

	apiURL := service.ServerURL + configProcessedSuffix
	apiURL = fmt.Sprintf(apiURL, service.AgentID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
