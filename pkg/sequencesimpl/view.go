/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencesimpl

import (
	"sync"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/appdef/sys"
	"github.com/voedger/voedger/pkg/istructs"
)

type view struct {
	vr      istructs.IViewRecords
	pid     istructs.PartitionID
	changes chan *map[istructs.WSID]map[appdef.QName]uint64
	wg      sync.WaitGroup
	keys    sync.Pool
	values  sync.Pool
}

func newView(vr istructs.IViewRecords, pid istructs.PartitionID) *view {
	return &view{
		vr:      vr,
		pid:     pid,
		changes: make(chan *map[istructs.WSID]map[appdef.QName]uint64),
		keys:    sync.Pool{New: func() any { return vr.KeyBuilder(sys.SequencesView) }},
		values:  sync.Pool{New: func() any { return vr.NewValueBuilder(sys.SequencesView) }},
	}
}

func (a *view) apply(changes *map[istructs.WSID]map[appdef.QName]uint64) {
	a.wg.Wait() // wait for previous updateView to finish

	a.changes <- changes
	a.wg.Add(1)
	go a.update() // a.wg.Done() in updateView
}

func (a *view) key(wsid istructs.WSID, name appdef.QName) istructs.IKeyBuilder {
	k := a.keys.Get().(istructs.IKeyBuilder)
	k.PutInt64(sys.SequencesView_PID, int64(a.pid))
	k.PutInt64(sys.SequencesView_WSID, int64(wsid))
	k.PutQName(sys.SequencesView_Name, name)
	return k
}

func (a *view) update() {
	defer a.wg.Done()
	changes := <-a.changes

	cnt := 0
	for _, s := range *changes {
		cnt += len(s)
	}

	batch := make([]istructs.ViewKV, 0, cnt)
	for wsid, ss := range *changes {
		for n, val := range ss {
			batch = append(batch, istructs.ViewKV{
				Key:   a.key(wsid, n),
				Value: a.value(val),
			})
		}
	}

	// retry on error
	if err := a.vr.PutBatch(istructs.NullWSID, batch); err != nil {
		panic(err)
	}
}

func (a *view) value(last uint64) istructs.IValueBuilder {
	v := a.values.Get().(istructs.IValueBuilder)
	v.PutInt64(sys.SequencesView_Last, int64(last))
	return v
}
