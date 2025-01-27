/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovery

import "github.com/voedger/voedger/pkg/appdef"

func ProvidePRPView(adb appdef.IAppDefBuilder) {
	wsb := adb.AlterWorkspace(appdef.SysWorkspaceQName)

	view := wsb.AddView(prp_ViewName)
	view.Key().PartKey().AddField(prp_PID, appdef.DataKind_int64)
	view.Key().ClustCols().AddField(prp_WSID, appdef.DataKind_int64)
	view.Value().AddField(prp_Offset, appdef.DataKind_int64, true)
	view.Value().AddField(prp_BaseRecordID, appdef.DataKind_RecordID, false)
	view.Value().AddField(prp_CBaseRecordID, appdef.DataKind_RecordID, false)
}
