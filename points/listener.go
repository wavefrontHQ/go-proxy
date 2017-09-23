package points

import (
	"bufio"
	"fmt"
	"github.com/wavefronthq/go-proxy/api"
	"github.com/wavefronthq/go-proxy/points/decoder"
	"log"
	"net"
)

// Interface that handles listening for points
type PointListener interface {
	StartListener(numForwarders, flushInterval, bufferSize, maxFlushSize int, format, workUnitId string, service api.WavefrontAPI)
	StopListener()
}

type DefaultPointListener struct {
	Port      int
	Builder   decoder.DecoderBuilder
	ptHandler PointHandler
}

func (ptListener *DefaultPointListener) StartListener(numForwarders, flushInterval, bufferSize, maxFlushSize int,
	format, workUnitId string, service api.WavefrontAPI) {

	log.Printf("Starting listener on port: %d\n", ptListener.Port)

	ptListener.ptHandler = &DefaultPointHandler{name: fmt.Sprintf("%d", ptListener.Port)}
	ptListener.ptHandler.init(numForwarders, flushInterval, bufferSize, maxFlushSize, format, workUnitId, service)

	connStr := fmt.Sprintf(":%d", ptListener.Port)
	addr, err := net.ResolveTCPAddr("tcp", connStr)
	if err != nil {
		panic(err)
	}

	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	go ptListener.startServer(tcpListener)
	log.Printf("Configured %d forwarders for %s listener on port: %d\n", numForwarders, format, ptListener.Port)
}

func (ptListener *DefaultPointListener) startServer(tcpListener *net.TCPListener) {
	for {
		// Listen for incoming connections
		conn, err := tcpListener.Accept()
		if err != nil || conn == nil {
			log.Printf("%d-listener: error accepting connection: %v\n", ptListener.Port, err.Error())
			continue
		}

		// Handle connections in a new goroutine
		go ptListener.handleRequest(conn)
	}
}

// Handles incoming requests.
func (ptListener *DefaultPointListener) handleRequest(conn net.Conn) {
	var ptDecoder decoder.PointDecoder = ptListener.Builder.Build()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		pointBytes := scanner.Bytes()
		point, err := ptDecoder.Decode(pointBytes)
		if err != nil {
			log.Println("Error decoding point", err)
			ptListener.ptHandler.HandleBlockedPoint(string(pointBytes))
			continue
		}
		ptListener.ptHandler.ReportPoint(point)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("%d-listener: error during scan: %v\n", ptListener.Port, err)
	}
	conn.Close()
}

func (ptListener *DefaultPointListener) StopListener() {
	//TODO: gracefully shutdown TCP listener
	log.Println("Stopping listener", ptListener.Port)
	ptListener.ptHandler.stop()
}
