package client

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/fasthttp-contrib/websocket"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type WSClient struct {
	url     string
	timeout time.Duration
	conn    *websocket.Conn
	inbox   chan []byte
	ioErr   error
	closed  bool
	wg      *sync.WaitGroup
}

func NewWSClient(url string, to ...time.Duration) (*WSClient, error) {
	timeout := time.Second
	if len(to) > 0 {
		timeout = to[0]
	}
	wscli := &WSClient{url: url, timeout: timeout, inbox: make(chan []byte, 32), wg: &sync.WaitGroup{}}
	err := wscli.Connect()
	return wscli, err
}

var dialer *websocket.Dialer = &websocket.Dialer{
	Proxy:           http.ProxyFromEnvironment,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SetDialer(d *websocket.Dialer) {
	dialer = d
}

func (c *WSClient) Connect() error {
	conn, resp, err := dialer.Dial(c.url, nil)
	if err != nil {
		var b []byte = nil
		var status int = 0
		if resp != nil {
			b, _ = ioutil.ReadAll(resp.Body)
			status = resp.StatusCode
		}
		return errors.New(fmt.Sprintf("dial err at %s: %d '%s', %s", c.url, status, string(b), err.Error()))
	}
	c.conn = conn
	c.closed = false
	c.ioErr = nil
	c.wg.Add(1)
	go c.read()
	return nil
}

func (c *WSClient) Close() error {
	err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close"), time.Now().Add(c.timeout))
	if err != nil {
		return err
	}
	c.wg.Wait()
	return c.conn.Close()
}

func (c *WSClient) IsClosed() bool {
	return c.closed
}

func (c *WSClient) read() {
	for c.ioErr == nil {
		mtype, bytes, err := c.conn.ReadMessage()
		if err != nil {
			c.ioErr = errors.New(fmt.Sprintf("conn read: %s", err.Error()))
			break
		} else if mtype != websocket.BinaryMessage {
			c.ioErr = errors.New(fmt.Sprintf("Expected binary message but was %d", mtype))
			break
		} else {
			c.inbox <- bytes
		}
	}
	c.wg.Done()
	c.closed = true
}

func (c *WSClient) ReadMessage() ([]byte, error) {
	if c.ioErr != nil {
		return nil, errors.New(fmt.Sprintf("read message io error: %s", c.ioErr.Error()))
	}
	select {
	case msg := <-c.inbox:
		logrus.Debugf("Got message %d bytes", len(msg))
		return msg, nil
	case <-time.After(c.timeout):
		if c.ioErr != nil {
			return nil, errors.New(fmt.Sprintf("read message io error: %s", c.ioErr.Error()))
		}
		return nil, errors.New("Failed to read message before timeout")
	}
}

func (c *WSClient) WriteMessage(msg []byte) error {
	if c.ioErr != nil {
		return errors.New(fmt.Sprintf("ws write io error: %s", c.ioErr))
	}
	if c.timeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	err := c.conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		return errors.New(fmt.Sprintf("ws write error: %s", err.Error()))
	}
	return nil
}

// Writes a message to the active websocket and expects given response
func (c *WSClient) WriteAndReadResponseMessage(msg []byte) ([]byte, error) {
	err := c.WriteMessage(msg)
	if err != nil {
		return []byte{}, err
	}
	return c.ReadMessage()
}

