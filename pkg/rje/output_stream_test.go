package rje

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestOutputStreamMaintainsOrder(t *testing.T) {
	outputStream := NewOutputStream()
	defer outputStream.Close()
	
	var wg sync.WaitGroup

	asyncMethod := func(arr []string) {
		defer wg.Done()
		for i := 0; i < len(arr); i++ {
			if _, err := outputStream.Write([]byte(arr[i])); err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
		}
	}

	wg.Add(2)
	go asyncMethod([]string{"0", "2", "4", "6", "8"})
	time.Sleep(time.Second / 2)
	go asyncMethod([]string{"1", "3", "5", "7", "9"})
	wg.Wait()

	buffer := outputStream.GetBuffer()
	if !bytes.Equal(buffer, []byte("0123456789")) {
		t.Errorf("buffer isn't as expected: %s", string(buffer))
	}
}
