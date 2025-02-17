/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/sequences"
)

// Sequencer implements sequences.ISequencer
type Sequencer struct {
	ctx        context.Context
	pid        istructs.PartitionID
	str        istructs.IAppStructs
	status     atomic.Uint32
	flusher    *flusher
	plogOffset istructs.Offset
	ws         istructs.WSID
	wlogOffset map[istructs.WSID]istructs.Offset
	recID      map[istructs.WSID]istructs.RecordID
	cRecID     map[istructs.WSID]istructs.RecordID
	mu         sync.Mutex
}

// NewSequencer creates a new Sequencer instance
func NewSequencer(ctx context.Context, pid istructs.PartitionID, str istructs.IAppStructs) *Sequencer {
	return &Sequencer{
		ctx:        ctx,
		pid:        pid,
		str:        str,
		status:     atomic.Uint32{},
		flusher:    newFlusher(ctx, pid, str),
		wlogOffset: make(map[istructs.WSID]istructs.Offset),
		recID:      make(map[istructs.WSID]istructs.RecordID),
		cRecID:     make(map[istructs.WSID]istructs.RecordID),
	}
}

// StartEvent reserves offsets for PLog and WLog
func (s *Sequencer) StartEvent(wait time.Duration, ws istructs.WSID) (plogOffset, wlogOffset istructs.Offset) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.Status() {
	case sequences.Recover:
		select {
		case <-time.After(wait):
			if s.Status() != sequences.Ready {
				return istructs.NullOffset, istructs.NullOffset
			}
		case <-s.ctx.Done():
			return istructs.NullOffset, istructs.NullOffset
		}
		fallthrough
	case sequences.Ready:
		s.status.Store(uint32(sequences.Eventing))
		s.ws = ws
		s.plogOffset++
		s.wlogOffset[ws]++
		return s.plogOffset, s.wlogOffset[ws]
	default:
		panic("invalid state")
	}
}

// NextRecID returns the next available record ID for ODoc/ORecord/WDoc/WRecord
func (s *Sequencer) NextRecID() istructs.RecordID {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status() != sequences.Eventing {
		panic("invalid state")
	}
	s.recID[s.ws]++
	return s.recID[s.ws]
}

// NextCRecID returns the next available record ID for CDoc/CRecord
func (s *Sequencer) NextCRecID() istructs.RecordID {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status() != sequences.Eventing {
		panic("invalid state")
	}
	s.cRecID[s.ws]++
	return s.cRecID[s.ws]
}

// FinishEvent finalizes the current event
func (s *Sequencer) FinishEvent() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status() != sequences.Eventing {
		panic("invalid state")
	}
	s.flusher.collect(s.plogOffset, s.ws, s.wlogOffset[s.ws], s.recID[s.ws], s.cRecID[s.ws])
	s.status.Store(uint32(sequences.Ready))
}

// CancelEvent aborts the ongoing event
func (s *Sequencer) CancelEvent() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status() != sequences.Eventing {
		panic("invalid state")
	}
	s.recovery()
}

// Status returns the current SequenceStatus
func (s *Sequencer) Status() sequences.SequenceStatus {
	return sequences.SequenceStatus(s.status.Load())
}

// recovery handles resetting the Sequencer state
func (s *Sequencer) recovery() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status() != sequences.None && s.Status() != sequences.Eventing {
		panic("invalid state")
	}
	s.status.Store(uint32(sequences.Recover))

	go func() {
		for {
			if s.ctx.Err() != nil {
				return
			}

			s.flusher.waitForInactive()

			// reset in-memory counters
			s.plogOffset = 0
			s.wlogOffset = make(map[istructs.WSID]istructs.Offset)
			s.recID = make(map[istructs.WSID]istructs.RecordID)
			s.cRecID = make(map[istructs.WSID]istructs.RecordID)

			// read from sequences view
			err := s.str.ViewRecords().Read(s.ctx, appdef.SysSequencesView, func(vr istructs.ViewRecord) error {
				ws := istructs.WSID(vr.AsInt64(appdef.SysSequencesView_WSID))
				switch vr.AsQName(appdef.SysSequencesView_Name) {
				case appdef.SysPLogOffsetSeq:
					s.plogOffset = istructs.Offset(vr.AsInt64(appdef.SysSequencesView_Last))
				case appdef.SysWLogOffsetSeq:
					s.wlogOffset[ws] = istructs.Offset(vr.AsInt64(appdef.SysSequencesView_Last))
				case appdef.SysRecIDSeq:
					s.recID[ws] = istructs.RecordID(vr.AsInt64(appdef.SysSequencesView_Last))
				case appdef.SysCRecIDSeq:
					s.cRecID[ws] = istructs.RecordID(vr.AsInt64(appdef.SysSequencesView_Last))
				}
				return nil
			})
			if err != nil {
				continue
			}

			// read from PLog to update counters
			err = s.str.Events().ReadPLog(s.ctx, s.pid, s.plogOffset, istructs.ReadToTheEnd, func(event istructs.PLogEvent) error {
				ws := event.WSID()
				for _, cud := range event.CUDs() {
					if cud.IsNew() {
						switch appdef.Type(cud.QName()).Kind() {
						case appdef.TypeKind_CDoc, appdef.TypeKind_CRecord:
							if cud.ID().BaseRecordID() > s.cRecID[ws] {
								s.cRecID[ws] = cud.ID().BaseRecordID()
							}
						case appdef.TypeKind_ODoc, appdef.TypeKind_ORecord, appdef.TypeKind_WDoc, appdef.TypeKind_WRecord:
							if cud.ID().BaseRecordID() > s.recID[ws] {
								s.recID[ws] = cud.ID().BaseRecordID()
							}
						}
					}
				}
				for _, obj := range event.Argument().Objects() {
					if obj.ID().BaseRecordID() > s.recID[ws] {
						s.recID[ws] = obj.ID().BaseRecordID()
					}
				}
				return nil
			})
			if err != nil {
				continue
			}

			s.status.Store(uint32(sequences.Ready))
			return
		}
	}()
}

// wait finalizes or waits for recovery
func (s *Sequencer) wait() {
	if s.Status() == sequences.Recover {
		for s.Status() != sequences.Ready {
			time.Sleep(10 * time.Millisecond)
		}
	} else {
		s.flusher.flush()
		s.flusher.waitForInactive()
		s.status.Store(uint32(sequences.Finished))
	}
}
