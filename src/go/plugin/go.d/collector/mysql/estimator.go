package mysql

import (
	"container/list"
	"time"
)

type storageRecord struct {
	stored int64
	ts     time.Time
}

type retentionTimeEstimator struct {
	records  *list.List
	Capacity int64
}

func newRetentionTimeEstimator() *retentionTimeEstimator {
	return &retentionTimeEstimator{records: list.New()}
}

func (rte *retentionTimeEstimator) Add(stored int64, ts time.Time) {
	rte.AddRecord(storageRecord{stored, ts})
}

func (rte *retentionTimeEstimator) AddRecord(record storageRecord) {
	if rte.records.Len() == 0 || rte.records.Back().Value.(storageRecord).stored < record.stored {
		rte.records.PushBack(record)
	}

	for rte.records.Len() > 0 {
		e := rte.records.Front()
		oldest := e.Value.(storageRecord)
		if oldest.stored+rte.Capacity >= record.stored {
			break
		}
		rte.records.Remove(e)
	}
}

func (rte retentionTimeEstimator) Estimate(now time.Time) time.Duration {
	oe := rte.records.Front().Value.(storageRecord)
	return now.Sub(oe.ts)
}
