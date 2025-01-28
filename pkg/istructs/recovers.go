/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package istructs

// Partition recovers
type IRecovers interface {
	// Gets partition recovery point from storage
	Get(PartitionID) (IPartitionRecoveryPoint, error)

	// Puts partition recovery point to storage
	Put(IPartitionRecoveryPoint) error
}

// Partition recovery point
type IPartitionRecoveryPoint interface {
	// Partition ID
	PID() PartitionID

	// Offset of the last committed event
	PLogOffset() Offset

	// Workspaces, which are handled by this partition
	Workspaces() map[WSID]IWorkspaceRecoveryPoint

	// Updates partition recovery point with new values.
	// It is repeatedly called to update information about workspaces recovery points.
	//
	// If workspace recovery point with the same WSID not exists - it will be created, otherwise updated.
	Update(plog Offset, ws WSID, wlog Offset, id, cid RecordID)
}

// Workspace recovery point
type IWorkspaceRecoveryPoint interface {
	// Workspace ID
	WSID() WSID

	// Offset of the last committed event
	WLogOffset() Offset

	// Last base record ID
	BaseRecordID() RecordID

	// Last CDoc (CRecord) base record ID
	CBaseRecordID() RecordID
}
