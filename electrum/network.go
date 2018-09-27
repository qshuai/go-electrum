package electrum

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
)

const delim = byte('\n')

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNodeConnected  = errors.New("node already connected")
)

type Transport interface {
	SendMessage([]byte) error
	Responses() <-chan []byte
	Errors() <-chan error
}

type respMetadata struct {
	Id     uint64  `json:"id"`
	Method string  `json:"method"`
	Error  *APIErr `json:"error"`
}

type APIErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type request struct {
	Id     uint64        `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type basicResp struct {
	Result string `json:"result"`
}

type Node struct {
	transport Transport

	handlersLock sync.RWMutex
	handlers     map[uint64]chan []byte

	pushHandlersLock sync.RWMutex
	pushHandlers     map[string][]chan []byte

	// nextId tags a request, and get the same id from server result.
	// Should be atomic operation for concurrence.
	// notice the max request limit, if reach to the max times,
	// 0 will be the next id. Assume the oldest has been deal completely.
	nextId uint64
}

// NewNode creates a new node.
func NewNode() *Node {
	n := &Node{
		handlers:     make(map[uint64]chan []byte),
		pushHandlers: make(map[string][]chan []byte),
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
		case _ = <-n.transport.Errors():
			return
		case bytes := <-n.transport.Responses():
			msg := &respMetadata{}
			if err := json.Unmarshal(bytes, msg); err != nil {
				return
			}
			if msg.Error != nil {
				return
			}
			if len(msg.Method) > 0 {
				n.pushHandlersLock.RLock()
				handlers := n.pushHandlers[msg.Method]
				n.pushHandlersLock.RUnlock()

				for _, handler := range handlers {
					select {
					case handler <- bytes:
					default:
					}
				}
			}

			n.handlersLock.RLock()
			c, ok := n.handlers[msg.Id]
			n.handlersLock.RUnlock()

			if ok {
				c <- bytes
			}
		}
	}
}

// listenPush returns a channel of messages matching the method.
func (n *Node) listenPush(method string) <-chan []byte {
	c := make(chan []byte, 1)
	n.pushHandlersLock.Lock()
	defer n.pushHandlersLock.Unlock()
	n.pushHandlers[method] = append(n.pushHandlers[method], c)
	return c
}

// request makes a request to the server and unmarshals the response into v.
func (n *Node) request(method string, params []interface{}, v interface{}) error {
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

	c := make(chan []byte, 1)

	n.handlersLock.Lock()
	n.handlers[msg.Id] = c
	n.handlersLock.Unlock()

	// TODO: deal with block, for example: network breaks down
	resp := <-c

	n.handlersLock.Lock()
	defer n.handlersLock.Unlock()
	delete(n.handlers, msg.Id)

	if v != nil {
		if err := json.Unmarshal(resp, v); err != nil {
			return err
		}
	}

	return nil
}
