package dispatcher

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSetEventStreamCount(t *testing.T) {
	SetEventStreamCount(100)
	require.Equal(t, uint64(100), LoadEventStreamCount())
}

func TestAddEventStreamCount(t *testing.T) {
	old := LoadEventStreamCount()
	AddEventStreamCount()
	new := LoadEventStreamCount()
	require.Equal(t, old+1, new)
}

func TestAddEventStreamCountConcurrent(t *testing.T) {
	old := LoadEventStreamCount()
	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			AddEventStreamCount()
			wg.Done()
		}()
	}
	wg.Wait()

	new := LoadEventStreamCount()
	require.Equal(t, old+100, new)
}
