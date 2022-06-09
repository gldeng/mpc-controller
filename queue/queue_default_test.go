package queue

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueueDefault(t *testing.T) {
	t.Parallel()

	require.Equal(t, true, Empty())

	data := []int{0, 1, 2, 3, 4}
	for i := 0; i < 5; i++ {
		Enqueue(i)
	}
	require.Equal(t, 5, Count())
	require.Equal(t, 4, Peek())
	require.Equal(t, false, Empty())

	data_ := make([]int, 0, 5)
	for {
		e := Dequeue()
		if e == nil {
			break
		}
		data_ = append(data_, e.(int))
	}
	require.Equal(t, true, Empty())
	require.Equal(t, nil, Peek())
	require.Equal(t, 0, Count())
	require.Equal(t, data, data_)

	Enqueue("hello")
	Clear()
	require.Equal(t, true, Empty())
}
