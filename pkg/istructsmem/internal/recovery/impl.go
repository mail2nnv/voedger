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

func (r Recovers) Get(pid istructs.PartitionID, point *istructs.PartitionRecoveryPoint) error {
	kb := r.vr.KeyBuilder(prp_ViewName)
	kb.PutInt64(prp_PID, int64(pid))

	return r.vr.Read(context.Background(), 0, kb, func(key istructs.IKey, value istructs.IValue) error {
		switch ws := istructs.WSID(key.AsInt64(prp_WSID)); ws {
		case 0:
			point.PLogOffset = istructs.Offset(value.AsInt64(prp_Offset))
		default:
			point.Workspaces[ws] = istructs.WorkspaceRecoveryPoint{
				WLogOffset:    istructs.Offset(value.AsInt64(prp_Offset)),
				BaseRecordID:  value.AsRecordID(prp_BaseRecordID),
				CBaseRecordID: value.AsRecordID(prp_CBaseRecordID),
			}
		}
		return nil
	})
}

func (r Recovers) Put(pid istructs.PartitionID, point istructs.PartitionRecoveryPoint) error {

	wsKey := func(ws istructs.WSID) istructs.IKeyBuilder {
		kb := r.vr.KeyBuilder(prp_ViewName)
		kb.PutInt64(prp_PID, int64(pid))
		kb.PutInt64(prp_WSID, int64(ws))
		return kb
	}

	wsValue := func(ws istructs.WorkspaceRecoveryPoint) istructs.IValueBuilder {
		vb := r.vr.NewValueBuilder(prp_ViewName)
		vb.PutInt64(prp_Offset, int64(ws.WLogOffset))
		vb.PutRecordID(prp_BaseRecordID, ws.BaseRecordID)
		vb.PutRecordID(prp_CBaseRecordID, ws.CBaseRecordID)
		return vb
	}

	batch := make([]istructs.ViewKV, 0, len(point.Workspaces)+1)
	batch = append(batch, istructs.ViewKV{
		Key:   wsKey(0),
		Value: wsValue(istructs.WorkspaceRecoveryPoint{WLogOffset: point.PLogOffset}),
	})

	for ws, wsPoint := range point.Workspaces {
		batch = append(batch, istructs.ViewKV{
			Key:   wsKey(ws),
			Value: wsValue(wsPoint),
		})
	}

	return r.vr.PutBatch(0, batch)
}
