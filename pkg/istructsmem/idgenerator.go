/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 * @author Denis Gribanov
 */

package istructsmem

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
)

type implIIDGenerator struct {
	nextBaseID            istructs.RecordID
	nextCDocCRecordBaseID istructs.RecordID
	onNewID               func(rawID, storageID istructs.RecordID, t appdef.IType) error
}

// used in tests
func NewIDGeneratorWithHook(onNewID func(rawID, storageID istructs.RecordID, t appdef.IType) error) istructs.IIDGenerator {
	return &implIIDGenerator{
		nextBaseID:            istructs.FirstBaseRecordID,
		nextCDocCRecordBaseID: istructs.FirstBaseRecordID,
		onNewID:               onNewID,
	}
}

func NewIDGenerator() istructs.IIDGenerator {
	return NewIDGeneratorWithHook(nil)
}

func (g implIIDGenerator) LastBaseID(t appdef.TypeKind) istructs.RecordID {
	if t == appdef.TypeKind_CDoc || t == appdef.TypeKind_CRecord {
		return g.nextCDocCRecordBaseID - 1
	}
	return g.nextBaseID - 1
}

func (g *implIIDGenerator) NextID(rawID istructs.RecordID, t appdef.IType) (storageID istructs.RecordID, err error) {
	if t.Kind() == appdef.TypeKind_CDoc || t.Kind() == appdef.TypeKind_CRecord {
		storageID = istructs.NewCDocCRecordID(g.nextCDocCRecordBaseID)
		g.nextCDocCRecordBaseID++
	} else {
		storageID = istructs.NewRecordID(g.nextBaseID)
		g.nextBaseID++
	}
	if g.onNewID != nil {
		if err := g.onNewID(rawID, storageID, t); err != nil {
			return istructs.NullRecordID, err
		}
	}
	return storageID, nil
}

func (g *implIIDGenerator) Update(id istructs.RecordID, t appdef.TypeKind) {
	if id >= istructs.MinClusterRecordID {
		// syncID>=322680000000000 -> the id is cluster record ID, includes generator ID, not base record id
		return
	}
	switch t {
	case appdef.TypeKind_CDoc, appdef.TypeKind_CRecord:
		g.nextCDocCRecordBaseID = id + 1
	case appdef.TypeKind_ODoc, appdef.TypeKind_ORecord:
		if id >= g.nextBaseID {
			// we do not know the order the IDs were issued for ODoc with ORecords
			// so let's bump if syncID is actually next
			g.nextBaseID = id + 1
		}
	default: // WDoc, WRecord
		g.nextBaseID = id + 1
	}
}

func (g *implIIDGenerator) UpdateOnSync(syncID istructs.RecordID, t appdef.TypeKind) {
	if syncID < istructs.MinClusterRecordID {
		// syncID<322680000000000 -> consider the syncID is from an old template.
		// ignore IDs from external registers
		// see https://github.com/voedger/voedger/issues/688
		return
	}
	g.Update(syncID.BaseRecordID(), t)
}
