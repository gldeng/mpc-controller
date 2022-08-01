package dispatcher

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEventObjects_Sort(t *testing.T) {
	evtObjs := EventObjects{
		&EventObject{EventNo: 1, Event: "1"},
		&EventObject{EventNo: 3, Event: "3"},
		&EventObject{EventNo: 2, Event: "2"},
	}
	evtObjs.Sort()
	require.Equal(t, uint64(1), evtObjs[0].EventNo)
	require.Equal(t, uint64(2), evtObjs[1].EventNo)
	require.Equal(t, uint64(3), evtObjs[2].EventNo)
}
