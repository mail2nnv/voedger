/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sys

import "github.com/voedger/voedger/pkg/appdef"

// Sequences view
var (
	SequencesView      appdef.QName     = appdef.NewQName(appdef.SysPackage, "sequences")
	SequencesView_PID  appdef.FieldName = "pid"  // int64 partition id
	SequencesView_WSID appdef.FieldName = "wsid" // int64 workspace id
	SequencesView_Name appdef.FieldName = "name" // QName sequence name
	SequencesView_Last appdef.FieldName = "last" // int64 last value
	// # Overall record size:
	//	8+8+2+8 = 26 bytes.
	//
	// # Max records per partition:
	// 	Max Cassandra partition size is 20Mb.
	//	Maximum 20*1024*1024/26 ≈ 806’596 records per partition.
	//
	// # Max sequences per workspace:
	// 	Let we have 1’000 partitions.
	//	Maximum 806’596/1`000 = 806 sequences per workspace.
	//
	// # Max workspaces per partition:
	//	Let we have only system sequences:
	//	- sys.plog_ofs, (1)
	//	- sys.wlog_ofs, sys.rec_id, sys.crec_id. (3 per ws)
	//	Maximum (806’596 - 1) / 3 = 268’865 workspaces per partition.
)
