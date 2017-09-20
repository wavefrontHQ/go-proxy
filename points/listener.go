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
	StartListener(builder decoder.DecoderBuilder, numForwarders, flushInterval, bufferSize, maxFlushSize int,
		format, workUnitId string, service api.WavefrontAPI)
	StopListener()
}

type DefaultPointListener struct {
	Name      string
	Port      int
	ptHandler PointHandler
}

func (ptListener *DefaultPointListener) StartListener(builder decoder.DecoderBuilder, numForwarders, flushInterval,
	bufferSize, maxFlushSize int, format, workUnitId string, service api.WavefrontAPI) {

	log.Printf("Starting %sListener on port: %d\n", ptListener.Name, ptListener.Port)

	ptListener.ptHandler = &DefaultPointHandler{name: fmt.Sprintf("%sPointHandler", ptListener.Name)}
	ptListener.ptHandler.initialize(numForwarders, flushInterval, bufferSize, maxFlushSize,
		format, workUnitId, service)

	connStr := fmt.Sprintf(":%d", ptListener.Port)
	addr, err := net.ResolveTCPAddr("tcp", connStr)
	if err != nil {
		panic(err)
	}

	tcpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	go ptListener.startServer(tcpListener, builder, ptListener.ptHandler)
	log.Printf("Configured %d forwarders for %s listener on port: %d\n", numForwarders, format, ptListener.Port)
}

func (ptListener *DefaultPointListener) startServer(tcpListener *net.TCPListener, builder decoder.DecoderBuilder, handler PointHandler) {
	for {
		// Listen for incoming connections
		conn, err := tcpListener.Accept()
		if err != nil || conn == nil {
			log.Println("Error accepting connection:", err.Error())
			continue
		}

		// Handle connections in a new goroutine
		go ptListener.handleRequest(conn, builder, handler)
	}
}

// Handles incoming requests.
func (ptListener *DefaultPointListener) handleRequest(conn net.Conn, builder decoder.DecoderBuilder, handler PointHandler) {

	var ptDecoder decoder.PointDecoder = builder.Build()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		point, err := ptDecoder.Decode(scanner.Bytes())
		if err != nil {
			fmt.Println("Error decoding point", err)
			continue
		}
		handler.ReportPoint(point)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Listener: error during scan: %v", err)
	}
	log.Println("Listener: closing connection request")
	conn.Close()
}

func (ptListener *DefaultPointListener) StopListener() {
	//TODO: gracefully shutdown TCP listener
	log.Println("Stopping listener", ptListener.Port)
	ptListener.ptHandler.shutdown()
}
