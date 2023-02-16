package core

import "time"

const (
	TxTypeAddDelegator = "AddDelegator"
	TxTypeExportC      = "ExportC"
	TxTypeImportP      = "ImportP"
)

type Parameters struct {
	IndexerStartDelay   time.Duration
	IndexerLoopDuration time.Duration
	// ReportFailureDelay has to be longer than IndexerLoopDuration so that we won't miss any tx when checking
	ReportFailureDelay       time.Duration
	TxIndexPurgueAge         time.Duration
	StakeTaskTimeoutDuration time.Duration
	UtxoBucketSeconds        int64
	EventLogChanCapacity     int
	QueueBufferChanCapacity  int
}

var (
	DefaultParameters *Parameters
)

func init() {
	setDefaultParameters()
	//setDefaultParametersInTestingMode() // Usage Note: only enable this during testing
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
