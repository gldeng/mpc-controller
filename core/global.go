package core

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	TxTypeAddDelegator = "AddDelegator"
	TxTypeExportC      = "ExportC"
	TxTypeImportP      = "ImportP"
)

// Parameters of the backend. Not configurable at node level so that all nodes share the same setting.
type Parameters struct {
	// After IndexerStartDelay period, the indexer will start to poll UTXOs.
	IndexerStartDelay time.Duration
	// IndexerLoopDuration indicates how frequently the indexer will poll UTXOs.
	IndexerLoopDuration time.Duration
	// ReportFailureDelay has to be longer than IndexerLoopDuration so that we won't miss any tx when checking
	ReportFailureDelay time.Duration
	// TxIndexPurgueAge indicates the threshold where older tx records in the index will be purged to free up memory.
	TxIndexPurgueAge time.Duration
	// StakeTaskTimeoutDuration indicates the timeout duration of stake task.
	StakeTaskTimeoutDuration time.Duration
	// UTXOs are grouped into buckets of UtxoBucketSeconds based the stake endTime so that all the principals / rewards
	// received in the same time window can be moved at one go. All nodes have  the same bucket definition to have
	// deterministic request definition.
	UtxoBucketSeconds       int64
	EventLogChanCapacity    int
	QueueBufferChanCapacity int
}

var (
	DefaultParameters *Parameters
)

func init() {
	setDefaultParameters()
	devMode := os.Getenv("DEV_MODE")
	if strings.Contains("YES_TRUE", strings.ToUpper(strings.TrimSpace(devMode))) {
		fmt.Println("devmode")
		setDefaultParametersInTestingMode() // Note: for testing locally, we loop faster
	}
}

func setDefaultParameters() {
	DefaultParameters = &Parameters{
		IndexerStartDelay:        1 * time.Minute,
		IndexerLoopDuration:      60 * time.Minute,
		ReportFailureDelay:       120 * time.Minute,
		TxIndexPurgueAge:         720 * time.Hour,
		StakeTaskTimeoutDuration: 5 * time.Hour,
		UtxoBucketSeconds:        2 * 60 * 60, // 2 hours
		EventLogChanCapacity:     1024,
		QueueBufferChanCapacity:  2048,
	}
}

func setDefaultParametersInTestingMode() {
	DefaultParameters = &Parameters{
		IndexerStartDelay:        1 * time.Minute,
		IndexerLoopDuration:      1 * time.Minute, // Set so that TxIndexer loops more frequently
		ReportFailureDelay:       120 * time.Minute,
		TxIndexPurgueAge:         720 * time.Hour,
		StakeTaskTimeoutDuration: 5 * time.Hour,
		UtxoBucketSeconds:        2 * 60 * 60, // Adjust it so that we can test more frequently
		EventLogChanCapacity:     1024,
		QueueBufferChanCapacity:  2048,
	}
}
