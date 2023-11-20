package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/juju/ratelimit"
	"github.com/kevwan/tproxy/display"
	"github.com/kevwan/tproxy/protocol"
)

const (
	useOfClosedConn = "use of closed network connection"
	statInterval    = time.Second * 5
)

var (
	errClientCanceled = errors.New("client canceled")
	stat              Stater
)

type PairedConnection struct {
	id       int
	cliConn  net.Conn
	svrConn  net.Conn
	once     sync.Once
	stopChan chan struct{}
}

func NewPairedConnection(id int, cliConn net.Conn) *PairedConnection {
	return &PairedConnection{
		id:       id,
		cliConn:  cliConn,
		stopChan: make(chan struct{}),
	}
}

func (c *PairedConnection) copyData(dst io.Writer, src io.Reader, tag string) {
	_, e := io.Copy(dst, src)
	if e != nil && e != io.EOF {
		netOpError, ok := e.(*net.OpError)
		if ok && netOpError.Err.Error() != useOfClosedConn {
			reason := netOpError.Unwrap().Error()
			display.PrintlnWithTime(color.HiRedString("[%d] %s error, %s", c.id, tag, reason))
		}
	}
}

func (c *PairedConnection) copyDataWithRateLimit(dst io.Writer, src io.Reader, tag string, limit int64) {
	if limit > 0 {
		bucket := ratelimit.NewBucket(time.Second, limit)
		src = ratelimit.Reader(src, bucket)
	}

	c.copyData(dst, src, tag)
}

func (c *PairedConnection) handleClientMessage() {
	// client closed also trigger server close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(c.svrConn, w)
	go protocol.CreateInterop(s.Protocol).Dump(r, protocol.ClientSide, c.id, s.Quiet)
	c.copyDataWithRateLimit(tee, c.cliConn, protocol.ClientSide, s.UpLimit)
}

func (c *PairedConnection) handleServerMessage() {
	// server closed also trigger client close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(newDelayedWriter(c.cliConn, s.Delay, c.stopChan), w)
	go protocol.CreateInterop(s.Protocol).Dump(r, protocol.ServerSide, c.id, s.Quiet)
	c.copyDataWithRateLimit(tee, c.svrConn, protocol.ServerSide, s.DownLimit)
}

func (c *PairedConnection) process(remote string) {
	defer c.stop()

	conn, err := net.Dial("tcp", remote)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("[x][%d] Couldn't connect to server: %v", c.id, err))
		return
	}

	display.PrintlnWithTime(color.HiGreenString("[%d] Connected to server: %s", c.id, conn.RemoteAddr()))

	stat.AddConn(strconv.Itoa(c.id), conn.(*net.TCPConn))
	c.svrConn = conn
	go c.handleServerMessage()

	c.handleClientMessage()
}

func (c *PairedConnection) stop() {
	c.once.Do(func() {
		close(c.stopChan)
		stat.DelConn(strconv.Itoa(c.id))

		if c.cliConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Client connection closed", c.id))
			c.cliConn.Close()
		}
		if c.svrConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Server connection closed", c.id))
			c.svrConn.Close()
		}
	})
}

func startListener() error {
	stat = NewStater(NewConnCounter(), NewStatPrinter(statInterval))
	go stat.Start()
	for k, v := range s.Mapping {
		go startOne(k, v)
	}
	select {}

}

func startOne(port uint32, remote string) {
	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.LocalHost, port))
	if err != nil {
		log.Panic(fmt.Errorf("failed to start listener: %w", err))
	}
	defer conn.Close()

	display.PrintfWithTime("Listening on %s...\n", conn.Addr().String())

	var connIndex int
	for {
		cliConn, err := conn.Accept()
		if err != nil {
			fmt.Printf("%+remote", fmt.Errorf("server: accept: %w", err))
		}

		connIndex++
		display.PrintlnWithTime(color.HiGreenString("[%d] Accepted from: %s",
			connIndex, cliConn.RemoteAddr()))

		pconn := NewPairedConnection(connIndex, cliConn)
		go pconn.process(remote)
	}
}
