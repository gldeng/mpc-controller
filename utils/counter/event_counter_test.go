package counter

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestStore(t *testing.T) {
	SetEventCount(100)
	require.Equal(t, uint64(100), LoadEventCount())
}

func TestAdd(t *testing.T) {
	old := LoadEventCount()
	AddEventCount()
	new := LoadEventCount()
	require.Equal(t, old+1, new)
}

func TestAddConcurrent(t *testing.T) {
	old := LoadEventCount()
	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			AddEventCount()
			wg.Done()
		}()
	}
	wg.Wait()

	new := LoadEventCount()
	require.Equal(t, old+100, new)
}
