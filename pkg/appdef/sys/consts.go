/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sys

import "github.com/voedger/voedger/pkg/appdef"

// Sequences view
var (
	// Sequences view name
	SequencesView appdef.QName = appdef.NewQName(appdef.SysPackage, "Sequences")

	// Sequence view fields
	SequencesView_PID  appdef.FieldName = "PID"  // int64 partition id, pk
	SequencesView_WSID appdef.FieldName = "WSID" // int64 workspace id, cc
	SequencesView_Name appdef.FieldName = "Seq"  // QName sequence name, cc
	SequencesView_Last appdef.FieldName = "Last" // int64 last value, value

	// Sequence names
	PlogOffsetSeq appdef.QName = appdef.NewQName(appdef.SysPackage, "PLogOffsetSeq")
	WlogOffsetSeq appdef.QName = appdef.NewQName(appdef.SysPackage, "WLogOffsetSeq")
	RecIDSeq      appdef.QName = appdef.NewQName(appdef.SysPackage, "RecIDSeq")
	CRecIDSeq     appdef.QName = appdef.NewQName(appdef.SysPackage, "CRecIDSeq")

	// # Overall record size:
	//	8+8+2+8 = 26 bytes.
	//
	// # Max records per partition:
	// 	Max Cassandra partition size is 20Mb.
	//	Maximum 20*1024*1024 / 26 = 806’596 records per partition.
	//
	// # Max sequences per workspace:
	// 	Let we have 1’000 partitions.
	//	Maximum 806’596 / 1`000 = 806 sequences per workspace.
	//
	// # Max workspaces per partition:
	//	Let we have only system sequences:
	//	- sys.plogOffsetSeq, (1)
	//	- sys.wlogOffsetSeq, sys.recIDSeq, sys.cRecIDSeq. (3 per ws)
	//	Maximum (806’596 - 1) / 3 = 268’865 workspaces per partition.
)
