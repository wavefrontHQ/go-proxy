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
	Start(numForwarders, flushInterval, bufferSize, maxFlushSize int, format, workUnitId string, service api.WavefrontAPI)
	Stop()
}

type DefaultPointListener struct {
	Port    int
	Builder decoder.DecoderBuilder
	handler PointHandler
}

func (l *DefaultPointListener) Start(numForwarders, flushInterval, bufferSize, maxFlushSize int,
	format, workUnitId string, service api.WavefrontAPI) {

	log.Printf("Starting listener on port: %d\n", l.Port)

	l.handler = &DefaultPointHandler{name: fmt.Sprintf("%d", l.Port)}
	l.handler.init(numForwarders, flushInterval, bufferSize, maxFlushSize, format, workUnitId, service)

	connStr := fmt.Sprintf(":%d", l.Port)
	addr, err := net.ResolveTCPAddr("tcp", connStr)
	if err != nil {
		panic(err)
	}

	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	go l.startServer(tcpListener)
	log.Printf("Configured %d forwarders for %s listener on port: %d\n", numForwarders, format, l.Port)
}

func (l *DefaultPointListener) startServer(tcpListener *net.TCPListener) {
	for {
		// Listen for incoming connections
		conn, err := tcpListener.Accept()
		if err != nil || conn == nil {
			log.Printf("%d-listener: error accepting connection: %v\n", l.Port, err.Error())
			continue
		}

		// Handle connections in a new goroutine
		go l.handleRequest(conn)
	}
}

// Handles incoming requests.
func (l *DefaultPointListener) handleRequest(conn net.Conn) {
	var pd decoder.PointDecoder = l.Builder.Build()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		pointBytes := scanner.Bytes()
		point, err := pd.Decode(pointBytes)
		if err != nil {
			log.Println("Error decoding point", err)
			l.handler.HandleBlockedPoint(string(pointBytes))
			continue
		}
		l.handler.ReportPoint(point)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("%d-listener: error during scan: %v\n", l.Port, err)
	}
	conn.Close()
}

func (l *DefaultPointListener) Stop() {
	//TODO: gracefully shutdown TCP listener
	log.Println("Stopping listener", l.Port)
	l.handler.stop()
}
