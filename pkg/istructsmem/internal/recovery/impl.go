/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovery

import (
	"context"

	"github.com/voedger/voedger/pkg/istructs"
)

// # Supports:
//   - istructs.IRecovers
type Recovers struct {
	vr istructs.IViewRecords
}

func NewRecovers(vr istructs.IViewRecords) *Recovers {
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
	modified   map[istructs.WSID]*WorkspaceRecoveryPoint
}

func NewPartitionRecoveryPoint(vr istructs.IViewRecords, pid istructs.PartitionID) *PartitionRecoveryPoint {
	p := &PartitionRecoveryPoint{
		vr:         vr,
		k:          vr.KeyBuilder(prp_ViewName),
		v:          vr.NewValueBuilder(prp_ViewName),
		pid:        pid,
		plog:       istructs.NullOffset,
		workspaces: make(map[istructs.WSID]istructs.IWorkspaceRecoveryPoint),
		modified:   make(map[istructs.WSID]*WorkspaceRecoveryPoint),
	}
	p.k.PutInt64(prp_PID, int64(pid))
	p.k.PutInt64(prp_WSID, int64(istructs.NullWSID))
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
		ws, ok := p.workspaces[wsID]
		if !ok {
			ws := NewWorkspaceRecoveryPoint(p.vr, p.PID(), wsID)
			p.workspaces[wsID] = ws
		}
		wsp := ws.(*WorkspaceRecoveryPoint)
		wsp.update(wlog, id, cid)
		p.modified[wsID] = wsp
	}
}

func (p *PartitionRecoveryPoint) get() error {
	clear(p.modified)
	clear(p.workspaces)

	k := p.vr.KeyBuilder(prp_ViewName)
	k.PutInt64(prp_PID, int64(p.pid))

	return p.vr.Read(context.Background(), istructs.NullWSID, k, func(key istructs.IKey, value istructs.IValue) error {
		ofs := istructs.Offset(value.AsInt64(prp_Offset))
		switch wsID := istructs.WSID(key.AsInt64(prp_WSID)); wsID {
		case istructs.NullWSID:
			p.plog = ofs
		default:
			ws := NewWorkspaceRecoveryPoint(p.vr, p.PID(), wsID)
			ws.update(
				ofs,
				value.AsRecordID(prp_BaseRecordID),
				value.AsRecordID(prp_CBaseRecordID),
			)
			p.workspaces[wsID] = ws
		}
		return nil
	})
}

func (p PartitionRecoveryPoint) key() istructs.IKeyBuilder { return p.k }

func (p *PartitionRecoveryPoint) put() (err error) {
	batch := make([]istructs.ViewKV, 0, len(p.modified)+1)
	batch = append(batch, istructs.ViewKV{
		Key:   p.key(),
		Value: p.value(),
	})

	for _, ws := range p.modified {
		batch = append(batch, istructs.ViewKV{
			Key:   ws.key(),
			Value: ws.value(),
		})
	}
	if err = p.vr.PutBatch(istructs.NullWSID, batch); err == nil {
		clear(p.modified)
	}
	return err
}

func (p *PartitionRecoveryPoint) value() istructs.IValueBuilder {
	p.v.PutInt64(prp_Offset, int64(p.plog))
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
		k:  vr.KeyBuilder(prp_ViewName),
		v:  vr.NewValueBuilder(prp_ViewName),
		ws: ws,
	}
	w.k.PutInt64(prp_PID, int64(pid))
	w.k.PutInt64(prp_WSID, int64(ws))
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
	w.v.PutInt64(prp_Offset, int64(w.wlog))
	w.v.PutRecordID(prp_BaseRecordID, w.id)
	w.v.PutRecordID(prp_CBaseRecordID, w.cid)
	return w.v
}
