package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/wavefronthq/go-proxy/agent"
	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/points"
	"github.com/wavefronthq/go-proxy/points/decoder"
)

var version = "0.1"

// flags
var fCfgPtr = flag.String("file", "", "Proxy configuration file")
var fTokenPtr = flag.String("token", "", "Wavefront API token")
var fServerPtr = flag.String("server", "", "Wavefront Server URL")
var fHostnamePtr = flag.String("host", "", "Hostname for the agent. Defaults to machine hostname")
var fWavefrontPortsPtr = flag.String("pushListenerPorts", "3878",
	"Comma-separated list of ports to listen on for Wavefront formatted data. Defaults to 2878.")
var fFlushThreadsPtr = flag.Int("flushThreads", 2, "Number of threads that flush to the server.")
var fFlushIntervalPtr = flag.Int("pushFlushInterval", 1000, "Milliseconds between flushes to the Wavefront server. Typically 1000.")
var fFlushMaxPointsPtr = flag.Int("pushFlushMaxPoints", 40000, "Max points per flush. Typically 40000.")
var fMaxBufferSizePtr = flag.Int("pushMemoryBufferLimit", 640000, "Max points to retain in memory. Defaults to 640000.")
var fIdFilePtr = flag.String("idFile", ".wavefront_id", "The agentId file.")
var fLogFilePtr = flag.String("logFile", "", "Output log file")

var listeners []points.PointListener

func parseFile(filename string) {
	//TODO: make config file driven
}

func waitForShutdown() {
	for {
		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt)
		select {
		case sig := <-signals:
			if sig == os.Interrupt {
				log.Println("Stopping Wavefront Proxy")
				stopListeners()
				os.Exit(0)
			}
		}
	}
}

func stopListeners() {
	for _, listener := range listeners {
		listener.Stop()
	}
}

func checkRequiredFlag(val string, msg string) {
	if val == "" {
		log.Println(msg)
		flag.Usage()
		os.Exit(1)
	}
}

func checkHostname() {
	if *fHostnamePtr == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("Error resolving hostname")
		}
		fHostnamePtr = &hostname
	}
}

func setupLogger() {
	if *fLogFilePtr != "" {
		f, err := os.Create(*fLogFilePtr)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}
}

func checkFlags() {
	flag.Parse()

	if *fCfgPtr != "" {
		parseFile(*fCfgPtr)
		return
	}
	checkRequiredFlag(*fTokenPtr, "Missing token")
	checkRequiredFlag(*fServerPtr, "Missing server")
	checkHostname()
	setupLogger()
}

func startListener(listener points.PointListener, service api.WavefrontAPI) {
	listener.Start(*fFlushThreadsPtr, *fFlushIntervalPtr, *fMaxBufferSizePtr, *fFlushMaxPointsPtr,
		api.FORMAT_GRAPHITE_V2, api.GRAPHITE_BLOCK_WORK_UNIT, service)
}

func startListeners(service api.WavefrontAPI) {
	ports := strings.Split(*fWavefrontPortsPtr, ",")
	for _, portStr := range ports {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal("Invalid port " + portStr)
		}
		listener := &points.DefaultPointListener{Port: port, Builder: decoder.GraphiteBuilder{}}
		listeners = append(listeners, listener)
		startListener(listener, service)
	}
}

func initAgent(agentID string, service api.WavefrontAPI) {
	agent := &agent.DefaultAgent{AgentID: agentID, ApiService: service}
	agent.InitAgent()
}

func main() {
	checkFlags()

	log.Println("Starting Wavefront Proxy")

	agentID := agent.CreateOrGetAgentId(*fIdFilePtr)
	apiService := &api.WavefrontAPIService{
		ServerURL: *fServerPtr,
		AgentID:   agentID,
		Hostname:  *fHostnamePtr,
		Token:     *fTokenPtr,
		Version:   version,
	}

	initAgent(agentID, apiService)
	startListeners(apiService)
	waitForShutdown()
}
