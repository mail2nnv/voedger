/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package istructs

// Partition recovers
type IRecovers interface {
	Get(PartitionID, *PartitionRecoveryPoint) error
	Put(PartitionID, PartitionRecoveryPoint) error
}

// Partition recovery point
type PartitionRecoveryPoint struct {
	// Offset of the last committed event
	PLogOffset Offset

	// Workspaces, which are handled by this partition
	Workspaces map[WSID]WorkspaceRecoveryPoint
}

// WorkspaceRecoveryPoint
type WorkspaceRecoveryPoint struct {
	// Offset of the last committed event
	WLogOffset Offset

	// Last base record ID
	BaseRecordID RecordID

	// Last CDoc (CRecord) base record ID
	CBaseRecordID RecordID
}
