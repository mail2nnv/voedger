/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencer

package sequencer
 
import (
		"context"
		"sync"
		"sync/atomic"
		"time"

		"github.com/voedger/voedger/pkg/appdef"
		"github.com/voedger/voedger/pkg/istructs"
)

type flusher struct {
		ctx     context.Context
		pid     istructs.PartitionID
		str     istructs.IAppStructs
		active  atomic.Bool
		changes *changes
		channel chan *changes
		keys    sync.Pool
		values  sync.Pool
}

const flushChangesCount = 100

func newFlusher(ctx context.Context, pid istructs.PartitionID, str istructs.IAppStructs) *flusher {
		fl := &flusher{
				ctx:     ctx,
				pid:     pid,
				str:     str,
				channel: make(chan *changes),
				keys: sync.Pool{
						New: func() interface{} { return str.ViewRecords().NewKeyBuilder(appdef.SysSequencesView) },
				},
				values: sync.Pool{
						New: func() interface{} { return str.ViewRecords().NewValueBuilder(appdef.SysSequencesView) },
				},
		}
		go fl.run()
		return fl
}

func (f *flusher) collect(plogOffset istructs.Offset, ws istructs.WSID, wlogOffset istructs.Offset, recID, cRecID istructs.RecordID) {
		if f.changes == nil {
				f.changes = &changes{
						wlogOffset: make(map[istructs.WSID]istructs.Offset),
						recID:      make(map[istructs.WSID]istructs.RecordID),
						cRecID:     make(map[istructs.WSID]istructs.RecordID),
				}
		}
		f.changes.plogOffset = plogOffset
		f.changes.wlogOffset[ws] = wlogOffset
		f.changes.recID[ws] = recID
		f.changes.cRecID[ws] = cRecID
		f.changes.count++
		if !f.active.Load() && f.changes.count >= flushChangesCount {
				f.flush()
		}
}

func (f *flusher) flush() {
		f.active.Store(true)
		f.channel <- f.changes
		f.changes = nil
}

func (f *flusher) run() {
		for changes := range f.channel {
				batchSize := len(changes.wlogOffset)*3 + 1
				batch := make([]istructs.ViewKV, 0, batchSize)

				// Build KV for each wlogOffset
				for ws, wlogOffset := range changes.wlogOffset {
						batch = append(batch, istructs.ViewKV{
								Key:   f.key(ws, appdef.SysWLogOffsetSeq),
								Value: f.value(wlogOffset),
						})
						if recID, ok := changes.recID[ws]; ok {
								batch = append(batch, istructs.ViewKV{
										Key:   f.key(ws, appdef.SysRecIDSeq),
										Value: f.value(recID),
								})
						}
						if cRecID, ok := changes.cRecID[ws]; ok {
								batch = append(batch, istructs.ViewKV{
										Key:   f.key(ws, appdef.SysCRecIDSeq),
										Value: f.value(cRecID),
								})
						}
				}

				// Build KV for plogOffset
				batch = append(batch, istructs.ViewKV{
						Key:   f.key(istructs.NullWSID, appdef.SysPLogOffsetSeq),
						Value: f.value(changes.plogOffset),
				})

				// Put batch
				for {
						err := f.str.ViewRecords().PutBatch(istructs.NullWSID, batch)
						if err == nil || f.ctx.Err() != nil {
								break
						}
				}

				f.active.Store(false)
		}
}

func (f *flusher) waitForInactive() {
		for f.active.Load() {
				time.Sleep(10 * time.Millisecond)
		}
}

func (f *flusher) key(ws istructs.WSID, seq appdef.QName) istructs.IKeyBuilder {
		key := f.keys.Get().(istructs.IKeyBuilder)
		key.PutInt64(appdef.SysSequencesView_PID, int64(f.pid))
		key.PutInt64(appdef.SysSequencesView_WSID, int64(ws))
		key.PutQName(appdef.SysSequencesView_Name, seq)
		return key
}

func (f *flusher) value(last uint64) istructs.IValueBuilder {
		val := f.values.Get().(istructs.IValueBuilder)
		val.PutInt64(appdef.SysSequencesView_Last, int64(last))
		return val
}
