package rje

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// Stream used to manage writing data to connected clients,
// and store job output to memory for future connecting clients
// Implements io.WriteCloser
type outputStream struct {
	buffer           *bytes.Buffer
	connectedWriters []io.Writer
	mutex            sync.RWMutex
	closed           bool
}

// Creates a new OutputStream for use by the library
func newOutputStream() *outputStream {
	return &outputStream{
		buffer:           bytes.NewBuffer([]byte{}),
		connectedWriters: []io.Writer{},
		mutex:            sync.RWMutex{},
		closed:           false,
	}
}

// OutputStream Write method writes the given data to its in memory buffer,
// as well as sending any data to all connected clients
func (oS *outputStream) Write(p []byte) (n int, err error) {
	oS.mutex.Lock()
	defer oS.mutex.Unlock()

	// Write to buffer first
	writeN, writeErr := oS.buffer.Write(p)

	// Write to all connected clients
	nilWriters := []int{}
	waitGroup := sync.WaitGroup{}

	for i, connectedWriter := range oS.connectedWriters {
		if connectedWriter != nil {
			waitGroup.Add(1)
			go func() {
				if _, err := connectedWriter.Write(p); err != nil {
					fmt.Println("Error sending data to connected client")
				}
				waitGroup.Done()
			}()
		} else {
			nilWriters = append(nilWriters, i)
		}
	}

	waitGroup.Wait()

	// Remove disconnected writers
	if len(nilWriters) > 0 {
		for i := len(nilWriters); i >= 0; i-- {
			oS.connectedWriters = append(oS.connectedWriters[:i], oS.connectedWriters[i+1:]...)
		}
	}

	return writeN, writeErr
}

// OutputStream GetBuffer returns the contents of the OutputStream
func (oS *outputStream) GetBuffer() []byte {
	oS.mutex.RLock()
	defer oS.mutex.RUnlock()

	var b []byte
	b = append(b, oS.buffer.Bytes()...)
	return b
}

func (oS *outputStream) IsClosed() bool {
	return oS.closed
}

// OutputStream Connect method adds a new client io.Writer to the list
// of connected clients
func (oS *outputStream) Connect(newWriter io.Writer) error {
	oS.mutex.Lock()
	defer oS.mutex.Unlock()
	oS.connectedWriters = append(oS.connectedWriters, newWriter)
	return nil
}

// OutputStream Close method to satisfy the io.WriterCloser interface
// Currently does nothing
func (oS *outputStream) Close() error {
	oS.closed = true
	return nil
}
