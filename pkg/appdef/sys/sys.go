/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sys

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/appdef/internal/datas"
)

func MakeSysPackage(adb appdef.IAppDefBuilder) {
	adb.AddPackage(appdef.SysPackage, appdef.SysPackagePath)

	makeSysWorkspace(adb)
}

func makeSysWorkspace(adb appdef.IAppDefBuilder) {
	wsb := adb.AddWorkspace(appdef.SysWorkspaceQName)
	ws := wsb.Workspace()

	// make sys data types
	for k := appdef.DataKind_null + 1; k < appdef.DataKind_FakeLast; k++ {
		_ = datas.NewSysData(ws, k)
	}

	// for projectors: sys.projectionOffsets
	viewProjectionOffsets := wsb.AddView(appdef.NewQName(appdef.SysPackage, "projectionOffsets"))
	viewProjectionOffsets.Key().PartKey().AddField("partition", appdef.DataKind_int32)
	viewProjectionOffsets.Key().ClustCols().AddField("projector", appdef.DataKind_QName)
	viewProjectionOffsets.Value().AddField("offset", appdef.DataKind_int64, true)

	// for child workspaces: sys.NextBaseWSID
	viewNextBaseWSID := wsb.AddView(appdef.NewQName(appdef.SysPackage, "NextBaseWSID"))
	viewNextBaseWSID.Key().PartKey().AddField("dummy1", appdef.DataKind_int32)
	viewNextBaseWSID.Key().ClustCols().AddField("dummy2", appdef.DataKind_int32)
	viewNextBaseWSID.Value().AddField("NextBaseWSID", appdef.DataKind_int64, true)

	// for sequences: sys.sequences
	viewPRP := wsb.AddView(SequencesView)
	viewPRP.Key().PartKey().AddField(SequencesView_PID, appdef.DataKind_int64)
	viewPRP.Key().ClustCols().AddField(SequencesView_WSID, appdef.DataKind_int64)
	viewPRP.Key().ClustCols().AddField(SequencesView_Name, appdef.DataKind_QName)
	viewPRP.Value().AddField(SequencesView_Last, appdef.DataKind_int64, false)
}
