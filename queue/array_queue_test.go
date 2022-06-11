package queue

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	type T struct {
		name  string
		queue *ArrayQueue
	}

	var ts = []*T{
		{"array queue", NewArrayQueue(64)},
	}

	for _, t_ := range ts {
		t_ := t_
		t.Run(t_.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, true, t_.queue.Empty())

			data := []int{0, 1, 2, 3, 4}
			for i := 0; i < 5; i++ {
				t_.queue.Enqueue(i)
			}
			require.Equal(t, 5, t_.queue.Count())
			require.Equal(t, 4, t_.queue.Peek())
			require.Equal(t, false, t_.queue.Empty())

			data_ := make([]int, 0, 5)
			for {
				e := t_.queue.Dequeue()
				if e == nil {
					break
				}
				data_ = append(data_, e.(int))
			}
			require.Equal(t, true, t_.queue.Empty())
			require.Equal(t, nil, t_.queue.Peek())
			require.Equal(t, 0, t_.queue.Count())
			require.Equal(t, data, data_)

			t_.queue.Enqueue("hello")
			t_.queue.Clear()
			require.Equal(t, true, t_.queue.Empty())
		})
	}
}
