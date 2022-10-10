package connection

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipeWrite(t *testing.T) {
	// conn write应该会写入channel
	service, client := net.Pipe()
	go func() {
		service.Write([]byte("hello world"))
		service.Close()
	}()
	msg, err := io.ReadAll(client)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(msg))

}
func TestPipeWriteThenRead(t *testing.T) {
	// conn write应该会写入channel
	server, client := net.Pipe()
	msgLength := 14
	client.Write([]byte("hello world"))

	go func() {
		// msg, err := io.ReadAll(client)
		msgB := make([]byte, msgLength)
		_, err := server.Read(msgB)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(string(msgB))

	}()

	// client.Close()
	time.Sleep(time.Second * 3)

}

func TestMConnectionSendFlushStop(t *testing.T) {

	server, client := net.Pipe()
	clientConn := CreateMConnection(client)
	clientConn.Start()
	msg := []byte("hello world")
	assert.True(t, clientConn.Send(0x01, msg))
	msgLength := 14
	errCh := make(chan error)
	// 接受数据
	go func() {
		msgB := make([]byte, msgLength)
		_, err := server.Read(msgB)
		fmt.Println(string(msgB))
		if err != nil {
			t.Error(err)
			return
		}
		errCh <- err
	}()

	// stop the conn - it should flush all conns
	clientConn.FlushStop()
	timer := time.NewTimer(3 * time.Second)
	select {
	case <-errCh:
	case <-timer.C:
		t.Error("timed out waiting for msgs to be read")
	}
}
