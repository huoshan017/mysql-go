package mysql_proxy_common

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"sync"
)

type ClientOnlyWrite struct {
	codec ClientCodecOnlyWrite

	reqMutex sync.Mutex // protects following
	request  Request

	mutex    sync.Mutex // protects following
	closing  bool       // user has called Close
	shutdown bool       // server has told us to stop
}

type ClientCodecOnlyWrite interface {
	WriteRequest(*Request, interface{}) error

	Close() error
}

func (client *ClientOnlyWrite) send(call *Call) error {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		client.mutex.Unlock()
		return errors.New("mysql_proxy: rpc client is shutdown or closing")
	}
	client.mutex.Unlock()

	// Encode and send the request.
	client.request.ServiceMethod = call.ServiceMethod
	err := client.codec.WriteRequest(&client.request, call.Args)
	return err
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
//
// The read and write halves of the connection are serialized independently,
// so no interlocking is required. However each half may be accessed
// concurrently so the implementation of conn should protect against
// concurrent reads or concurrent writes.
func NewClientOnlyWrite(conn io.ReadWriteCloser) *Client {
	encBuf := bufio.NewWriter(conn)
	client := &gobClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
	return NewClientWithCodec(client)
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodecOnlyWrite(codec ClientCodecOnlyWrite) *ClientOnlyWrite {
	client := &ClientOnlyWrite{
		codec: codec,
	}
	return client
}

// Dial connects to an RPC server at the specified network address.
func DialOnlyWrite(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	var buf = []byte{byte(CONNECTION_TYPE_WRITE)}
	_, err = conn.Write(buf)
	if err != nil {
		return nil, err
	}
	return NewClientOnlyWrite(conn), nil
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *ClientOnlyWrite) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.codec.Close()
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *ClientOnlyWrite) Call(serviceMethod string, args interface{}, reply interface{}) error {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	return client.send(call)
}
