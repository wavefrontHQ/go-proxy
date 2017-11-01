package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"fmt"
	"github.com/wavefronthq/go-proxy/agent"
	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/points"
	"github.com/wavefronthq/go-proxy/points/decoder"
)

// flags
var (
	fCfgPtr            = flag.String("file", "", "Proxy configuration file")
	fTokenPtr          = flag.String("token", "", "Wavefront API token")
	fServerPtr         = flag.String("server", "", "Wavefront Server URL")
	fHostnamePtr       = flag.String("host", "", "Hostname for the agent. Defaults to machine hostname")
	fWavefrontPortsPtr = flag.String("pushListenerPorts", "2878",
		"Comma-separated list of ports to listen on for Wavefront formatted data")
	fOpenTSDBPortsPtr = flag.String("opentsdbPorts", "4242",
		"Comma-separated list of ports to listen on for OpenTSDB formatted data")
	fFlushThreadsPtr   = flag.Int("flushThreads", 4, "Number of threads that flush to the server")
	fFlushIntervalPtr  = flag.Int("pushFlushInterval", 1000, "Milliseconds between flushes to the Wavefront server")
	fFlushMaxPointsPtr = flag.Int("pushFlushMaxPoints", 40000, "Max points per flush")
	fMaxBufferSizePtr  = flag.Int("pushMemoryBufferLimit", 640000, "Max points to retain in memory")
	fIdFilePtr         = flag.String("idFile", ".wavefront_id", "The agentId file")
	fLogFilePtr        = flag.String("logFile", "", "Output log file")
	fPprofAddr         = flag.String("pprof-addr", "", "pprof address to listen on, disabled if empty")
	fVersion           = flag.Bool("version", false, "Display the version and exit")
)

var (
	version   string
	commit    string
	branch    string
	tag       string
	listeners []points.PointListener
)

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

func getVersion() string {
	if tag == "" {
		return version
	}
	return tag
}

func checkFlags() {
	flag.Parse()

	// check for flags which do something and exit immediately
	switch {
	case *fVersion:
		fmt.Printf("wavefront-proxy v%s (git: %s %s)\n", getVersion(), branch, commit)
		os.Exit(0)
	}

	if *fCfgPtr != "" {
		parseFile(*fCfgPtr)
		return
	}
	checkRequiredFlag(*fTokenPtr, "Missing token")
	checkRequiredFlag(*fServerPtr, "Missing server")
	checkHostname()
	setupLogger()
}

func startPointListener(listener points.PointListener, service api.WavefrontAPI) {
	listener.Start(*fFlushThreadsPtr, *fFlushIntervalPtr, *fMaxBufferSizePtr, *fFlushMaxPointsPtr,
		api.FormatGraphiteV2, api.GraphiteBlockWorkUnit, service)
}

func startPointListeners(service api.WavefrontAPI, portsList string, builder decoder.DecoderBuilder) {
	ports := strings.Split(portsList, ",")
	for _, portStr := range ports {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal("Invalid port " + portStr)
		}
		listener := &points.DefaultPointListener{Port: port, Builder: builder}
		listeners = append(listeners, listener)
		startPointListener(listener, service)
	}
}

func startListeners(service api.WavefrontAPI) {
	if *fWavefrontPortsPtr != "" {
		startPointListeners(service, *fWavefrontPortsPtr, decoder.GraphiteBuilder{})
	}

	if *fOpenTSDBPortsPtr != "" {
		startPointListeners(service, *fOpenTSDBPortsPtr, decoder.OpenTSDBBuilder{})
	}
}

func initAgent(agentID, serverURL string, service api.WavefrontAPI) {
	agent := &agent.DefaultAgent{AgentID: agentID, ApiService: service, ServerURL: serverURL}
	agent.InitAgent()
}

func main() {
	checkFlags()

	log.Printf("Starting Wavefront Proxy Version %s", version)

	if *fPprofAddr != "" {
		go func() {
			log.Printf("Starting pprof HTTP server at: %s", *fPprofAddr)
			if err := http.ListenAndServe(*fPprofAddr, nil); err != nil {
				log.Fatal(err.Error())
			}
		}()
	}

	agentID := agent.CreateOrGetAgentId(*fIdFilePtr)
	apiService := &api.WavefrontAPIService{
		ServerURL: *fServerPtr,
		AgentID:   agentID,
		Hostname:  *fHostnamePtr,
		Token:     *fTokenPtr,
		Version:   version,
	}

	initAgent(agentID, *fServerPtr, apiService)
	startListeners(apiService)
	waitForShutdown()
}
