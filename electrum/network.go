package electrum

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const delim = byte('\n')

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNodeConnected  = errors.New("node already connected")
	ErrNodeShutdown   = errors.New("node has shutdown")
	ErrTimeout        = errors.New("request timeout")
)

type Transport interface {
	SendMessage([]byte) error
	Responses() <-chan []byte
	Errors() <-chan error
}

type response struct {
	Id     uint64  `json:"id"`
	Method string  `json:"method"`
	Error  *APIErr `json:"error"`
}

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIErr) Error() string {
	return fmt.Sprintf("errorNo: %d, errMsg: %s", e.Code, e.Message)
}

type request struct {
	Id     uint64        `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type basicResp struct {
	Result string `json:"result"`
}

type container struct {
	content []byte
	err     error
}

type Node struct {
	transport Transport

	handlersLock sync.RWMutex
	handlers     map[uint64]chan *container

	pushHandlersLock sync.RWMutex
	pushHandlers     map[string][]chan *container

	Error chan error
	quit  chan struct{}

	// nextId tags a request, and get the same id from server result.
	// Should be atomic operation for concurrence.
	// notice the max request limit, if reach to the max times,
	// 0 will be the next id. Assume the oldest has been deal completely.
	nextId uint64
}

// NewNode creates a new node.
func NewNode() *Node {
	n := &Node{
		handlers:     make(map[uint64]chan *container),
		pushHandlers: make(map[string][]chan *container),

		Error: make(chan error),
		quit:  make(chan struct{}),
	}

	return n
}

// ConnectTCP creates a new TCP connection to the specified address.
func (n *Node) ConnectTCP(addr string) error {
	if n.transport != nil {
		return ErrNodeConnected
	}

	transport, err := NewTCPTransport(addr)
	if err != nil {
		return err
	}
	n.transport = transport
	go n.listen()

	return nil
}

// ConnectSLL creates a new SLL connection to the specified address.
func (n *Node) ConnectSSL(addr string, config *tls.Config) error {
	if n.transport != nil {
		return ErrNodeConnected
	}
	transport, err := NewSSLTransport(addr, config)
	if err != nil {
		return err
	}
	n.transport = transport
	go n.listen()

	return nil
}

// listen processes messages from the server.
func (n *Node) listen() {
	for {
		select {
		case err := <-n.transport.Errors():
			n.Error <- err
			n.shutdown()
		case bytes := <-n.transport.Responses():
			result := &container{
				content: bytes,
			}

			msg := &response{}
			if err := json.Unmarshal(bytes, msg); err != nil {
				if DebugMode {
					log.Printf("unmarshal received message failed: %v", err)
				}

				result.err = fmt.Errorf("unmarshal received message failed: %v", err)
			} else {
				result.err = msg.Error
			}

			// subscribe message if returned message with 'method' field
			if len(msg.Method) > 0 {
				n.pushHandlersLock.RLock()
				handlers := n.pushHandlers[msg.Method]
				n.pushHandlersLock.RUnlock()

				for _, handler := range handlers {
					select {
					case handler <- result:
					default:
					}
				}
			}

			n.handlersLock.RLock()
			c, ok := n.handlers[msg.Id]
			n.handlersLock.RUnlock()

			if ok {
				c <- result
			}
		}
	}
}

// listenPush returns a channel of messages matching the method.
func (n *Node) listenPush(method string) <-chan *container {
	c := make(chan *container, 1)
	n.pushHandlersLock.Lock()
	n.pushHandlers[method] = append(n.pushHandlers[method], c)
	n.pushHandlersLock.Unlock()

	return c
}

// request makes a request to the server and unmarshals the response into v.
func (n *Node) request(method string, params []interface{}, v interface{}) error {
	select {
	case <-n.quit:
		return ErrNodeShutdown
	default:
	}

	msg := request{
		Id:     atomic.LoadUint64(&n.nextId),
		Method: method,
		Params: params,
	}
	atomic.AddUint64(&n.nextId, 1)
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	bytes = append(bytes, delim)
	if err := n.transport.SendMessage(bytes); err != nil {
		return err
	}

	c := make(chan *container, 1)

	n.handlersLock.Lock()
	n.handlers[msg.Id] = c
	n.handlersLock.Unlock()

	var resp *container
	select {
	case resp = <-c:
	case <-time.After(5 * time.Second):
		return ErrTimeout
	}

	if resp.err != nil {
		return resp.err
	}

	n.handlersLock.Lock()
	delete(n.handlers, msg.Id)
	n.handlersLock.Unlock()

	if v != nil {
		err = json.Unmarshal(resp.content, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) shutdown() {
	close(n.quit)

	n.transport = nil
	n.handlers = nil
	n.pushHandlers = nil
}
