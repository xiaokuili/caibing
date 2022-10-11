package connection

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	defaultSendTimeout = time.Second * 1
)

type MConnection struct {
	conn            net.Conn
	c               []byte
	bufConnWriter   *bufio.Writer
	bufConnReader   *bufio.Reader
	channel         *myChannel
	doneSendRoutine chan struct{}
	stopSendRoutine chan struct{}
}

func CreateMConnection(conn net.Conn) *MConnection {
	c := make([]byte, 0)
	ch := &myChannel{
		sendQueue: make(chan []byte),
		conn:      conn,
		sending:   nil,
	}
	doneSendRoutine := make(chan struct{})
	stopSendRoutine := make(chan struct{})
	c = append(c, 0x01)
	return &MConnection{
		conn:            conn,
		c:               c,
		bufConnWriter:   bufio.NewWriter(conn),
		bufConnReader:   bufio.NewReader(conn),
		channel:         ch,
		doneSendRoutine: doneSendRoutine,
		stopSendRoutine: stopSendRoutine,
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
FOR_LOOP:
	for {

		select {
		case <-c.stopSendRoutine:
			break FOR_LOOP
		default:
			c.sendSomePacketMsgs()
			time.Sleep(time.Microsecond * 100)
		}

	}
	close(c.doneSendRoutine)
}
func (c *MConnection) sendSomePacketMsgs() bool {

	// Now send some PacketMsgs.
	if c.sendPacketMsg() {
		return true
	}
	return false
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendPacketMsg() bool {
	c.channel.isSendPending()
	_, err := c.channel.writePacketMsgTo(c.conn)
	return err != nil
}

func (c *MConnection) FlushStop() {
	c.stopServices()
	<-c.doneSendRoutine
	c.sendSomePacketMsgs()
	c.conn.Close()
}
func (c *MConnection) stopServices() {
	close(c.stopSendRoutine)
}

type myChannel struct {
	sendQueue chan []byte
	conn      net.Conn
	sending   []byte
}

func (c *myChannel) isSendPending() {
	select {
	case msg := <-c.sendQueue:
		fmt.Println("sendQueue to sending")
		c.sending = msg
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

func (c *myChannel) writePacketMsgTo(w io.Writer) (n int, err error) {
	if c.sending != nil {
		fmt.Println("sending to conn")

		w.Write(c.sending)
		c.sending = nil
	}
	return 0, nil
}
