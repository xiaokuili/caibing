package connection

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

const (
	defaultSendTimeout = time.Second * 1
)

type MConnection struct {
	conn          net.Conn
	c             []byte
	bufConnWriter *bufio.Writer
	bufConnReader *bufio.Reader
	channel       *myChannel
}

func CreateMConnection(conn net.Conn) *MConnection {
	c := make([]byte, 0)
	ch := &myChannel{
		sendQueue: make(chan []byte),
		conn:      conn,
	}
	c = append(c, 0x01)
	return &MConnection{
		conn:          conn,
		c:             c,
		bufConnWriter: bufio.NewWriter(conn),
		bufConnReader: bufio.NewReader(conn),
		channel:       ch,
	}
}

func (c *MConnection) Start() {
	go c.sendRoutine()
}

func (c *MConnection) recvRoutine() {

}

func (c *MConnection) Send(chID byte, msg []byte) bool {
	for i := 0; i < len(c.c); i++ {
		if c.c[i] == chID {
			c.channel.sendBytes(msg)
			return true
		}
	}
	return false
}

func (c *MConnection) sendRoutine() {
	for {
		c.channel.isSendPending()
		time.Sleep(time.Microsecond * 100)
	}
}

func (c *MConnection) FlushStop() {
	err := c.bufConnWriter.Flush()
	if err != nil {
		panic(err)
	}
}

type myChannel struct {
	sendQueue chan []byte
	conn      net.Conn
}

func (c *myChannel) isSendPending() {
	select {
	case msg := <-c.sendQueue:
		fmt.Println("sendQueue to conn")
		c.conn.Write(msg)
	default:

	}
}

func (c *myChannel) sendBytes(bytes []byte) bool {
	fmt.Println("write to sendQueue")
	select {
	case c.sendQueue <- bytes:
		return true
	case <-time.After(defaultSendTimeout):
		return false
	}
}
