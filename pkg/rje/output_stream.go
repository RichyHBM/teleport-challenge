package rje

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// Implements io.WriteCloser
type OutputStream struct {
	buffer           *bytes.Buffer
	connectedWriters []io.Writer
	mutex            sync.RWMutex
}

func NewOutputStream() *OutputStream {
	return &OutputStream{
		buffer:           bytes.NewBuffer([]byte{}),
		connectedWriters: []io.Writer{},
		mutex:            sync.RWMutex{},
	}
}

func (oS *OutputStream) Write(p []byte) (n int, err error) {
	oS.mutex.Lock()
	defer oS.mutex.Unlock()

	// Write to all connected clients
	nilWriters := []int{}
	for i, connectedWriter := range oS.connectedWriters {
		if connectedWriter != nil {
			if _, err := connectedWriter.Write(p); err != nil {
				fmt.Println("Error sending data to connected client")
			}
		} else {
			nilWriters = append(nilWriters, i)
		}
	}

	// Remove disconnected writers
	if len(nilWriters) > 0 {
		for i := len(nilWriters); i >= 0; i-- {
			oS.connectedWriters = append(oS.connectedWriters[:i], oS.connectedWriters[i+1:]...)
		}
	}

	// Write to the in memory buffer, for future clients
	if oS.buffer.Available() < len(p) {
		oS.buffer.Grow(len(p) - oS.buffer.Available())
	}

	return oS.buffer.Write(p)
}

func (oS *OutputStream) GetBuffer() []byte {
	oS.mutex.RLock()
	defer oS.mutex.RUnlock()

	return oS.buffer.Bytes()
}

func (oS *OutputStream) Connect(newWriter io.Writer) error {
	oS.mutex.Lock()
	defer oS.mutex.Unlock()
	oS.connectedWriters = append(oS.connectedWriters, newWriter)
	return nil
}

func (oS *OutputStream) Close() error {
	return nil
}
