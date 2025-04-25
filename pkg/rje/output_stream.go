package rje

import (
	"bytes"
	"sync"
)

// Implements io.WriteCloser
type OutputStream struct {
	buffer *bytes.Buffer
	mutex  sync.RWMutex
}

func NewOutputStream() *OutputStream {
	return &OutputStream{
		buffer: bytes.NewBuffer([]byte{}),
		mutex:  sync.RWMutex{},
	}
}

func (oS *OutputStream) Write(p []byte) (n int, err error) {
	oS.mutex.Lock()
	defer oS.mutex.Unlock()

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

func (oS *OutputStream) Close() error {
	return nil
}
