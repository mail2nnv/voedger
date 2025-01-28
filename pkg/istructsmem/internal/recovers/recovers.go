/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovers

import (
	"context"

	"github.com/voedger/voedger/pkg/appdef/sys"
	"github.com/voedger/voedger/pkg/istructs"
)

// # Supports:
//   - istructs.IRecovers
type Recovers struct {
	vr istructs.IViewRecords
}

func New(vr istructs.IViewRecords) *Recovers {
	return &Recovers{
		vr: vr,
	}
}

func (r Recovers) Get(pid istructs.PartitionID) (istructs.IPartitionRecoveryPoint, error) {
	point := NewPartitionRecoveryPoint(r.vr, pid)

	if err := point.get(); err != nil {
		return nil, err
	}

	return point, nil
}

func (r Recovers) Put(point istructs.IPartitionRecoveryPoint) error {
	return point.(*PartitionRecoveryPoint).put()
}

// # Supports:
//   - istructs.IPartitionRecoveryPoint
type PartitionRecoveryPoint struct {
	vr         istructs.IViewRecords
	k          istructs.IKeyBuilder
	v          istructs.IValueBuilder
	pid        istructs.PartitionID
	plog       istructs.Offset
	workspaces map[istructs.WSID]istructs.IWorkspaceRecoveryPoint
	modified   map[istructs.WSID]istructs.IWorkspaceRecoveryPoint
}

func NewPartitionRecoveryPoint(vr istructs.IViewRecords, pid istructs.PartitionID) *PartitionRecoveryPoint {
	p := &PartitionRecoveryPoint{
		vr:         vr,
		k:          vr.KeyBuilder(sys.Prp_ViewName),
		v:          vr.NewValueBuilder(sys.Prp_ViewName),
		pid:        pid,
		plog:       istructs.NullOffset,
		workspaces: make(map[istructs.WSID]istructs.IWorkspaceRecoveryPoint),
		modified:   make(map[istructs.WSID]istructs.IWorkspaceRecoveryPoint),
	}
	p.k.PutInt64(sys.Prp_PID, int64(pid))
	p.k.PutInt64(sys.Prp_WSID, int64(istructs.NullWSID))
	return p
}

func (p PartitionRecoveryPoint) PID() istructs.PartitionID { return p.pid }

func (p PartitionRecoveryPoint) PLogOffset() istructs.Offset { return p.plog }

func (p PartitionRecoveryPoint) Workspaces() map[istructs.WSID]istructs.IWorkspaceRecoveryPoint {
	return p.workspaces
}

func (p *PartitionRecoveryPoint) Update(plog istructs.Offset, wsID istructs.WSID, wlog istructs.Offset, id, cid istructs.RecordID) {
	p.plog = plog

	if wsID != 0 {
		w, ok := p.workspaces[wsID]
		if !ok {
			w = NewWorkspaceRecoveryPoint(p.vr, p.PID(), wsID)
			p.workspaces[wsID] = w
		}
		w.(*WorkspaceRecoveryPoint).update(wlog, id, cid)
		p.modified[wsID] = w
	}
}

func (p *PartitionRecoveryPoint) get() error {
	clear(p.modified)
	clear(p.workspaces)

	k := p.vr.KeyBuilder(sys.Prp_ViewName)
	k.PutInt64(sys.Prp_PID, int64(p.pid))

	return p.vr.Read(context.Background(), istructs.NullWSID, k, func(key istructs.IKey, value istructs.IValue) error {
		ofs := istructs.Offset(value.AsInt64(sys.Prp_Offset))
		switch wsID := istructs.WSID(key.AsInt64(sys.Prp_WSID)); wsID {
		case istructs.NullWSID:
			p.plog = ofs
		default:
			w := NewWorkspaceRecoveryPoint(p.vr, p.PID(), wsID)
			w.update(
				ofs,
				value.AsRecordID(sys.Prp_BaseRecordID),
				value.AsRecordID(sys.Prp_CBaseRecordID),
			)
			p.workspaces[wsID] = w
		}
		return nil
	})
}

func (p PartitionRecoveryPoint) key() istructs.IKeyBuilder { return p.k }

func (p *PartitionRecoveryPoint) put() error {
	batch := make([]istructs.ViewKV, 0, len(p.modified)+1)
	batch = append(batch, istructs.ViewKV{
		Key:   p.key(),
		Value: p.value(),
	})

	for _, w := range p.modified {
		w := w.(*WorkspaceRecoveryPoint)
		batch = append(batch, istructs.ViewKV{
			Key:   w.key(),
			Value: w.value(),
		})
	}
	if err := p.vr.PutBatch(istructs.NullWSID, batch); err != nil {
		return err
	}
	clear(p.modified)
	return nil
}

func (p *PartitionRecoveryPoint) value() istructs.IValueBuilder {
	p.v.PutInt64(sys.Prp_Offset, int64(p.plog))
	return p.v
}

// # Supports:
//   - istructs.IWorkspaceRecoveryPoint
type WorkspaceRecoveryPoint struct {
	vr   istructs.IViewRecords
	k    istructs.IKeyBuilder
	v    istructs.IValueBuilder
	ws   istructs.WSID
	wlog istructs.Offset
	id   istructs.RecordID
	cid  istructs.RecordID
}

func NewWorkspaceRecoveryPoint(vr istructs.IViewRecords, pid istructs.PartitionID, ws istructs.WSID) *WorkspaceRecoveryPoint {
	w := &WorkspaceRecoveryPoint{
		vr: vr,
		k:  vr.KeyBuilder(sys.Prp_ViewName),
		v:  vr.NewValueBuilder(sys.Prp_ViewName),
		ws: ws,
	}
	w.k.PutInt64(sys.Prp_PID, int64(pid))
	w.k.PutInt64(sys.Prp_WSID, int64(ws))
	return w
}

func (w WorkspaceRecoveryPoint) WSID() istructs.WSID { return w.ws }

func (w WorkspaceRecoveryPoint) WLogOffset() istructs.Offset { return w.wlog }

func (w WorkspaceRecoveryPoint) BaseRecordID() istructs.RecordID { return w.id }

func (w WorkspaceRecoveryPoint) CBaseRecordID() istructs.RecordID { return w.cid }

func (w WorkspaceRecoveryPoint) key() istructs.IKeyBuilder { return w.k }

func (w *WorkspaceRecoveryPoint) update(wlog istructs.Offset, id, cid istructs.RecordID) {
	w.wlog = wlog
	w.id = id
	w.cid = cid
}

func (w *WorkspaceRecoveryPoint) value() istructs.IValueBuilder {
	w.v.PutInt64(sys.Prp_Offset, int64(w.wlog))
	w.v.PutRecordID(sys.Prp_BaseRecordID, w.id)
	w.v.PutRecordID(sys.Prp_CBaseRecordID, w.cid)
	return w.v
}
