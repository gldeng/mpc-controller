package misc

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewID(t *testing.T) {
	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			id := NewID()
			fmt.Println(id)
			wg.Done()
		}()
	}
	wg.Wait()
}
