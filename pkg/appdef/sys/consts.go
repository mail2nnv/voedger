/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sys

import "github.com/voedger/voedger/pkg/appdef"

// Partition recovery points view
var (
	Prp_ViewName      appdef.QName     = appdef.NewQName(appdef.SysPackage, "prpView")
	Prp_PID           appdef.FieldName = "pid"
	Prp_WSID          appdef.FieldName = "wsid"
	Prp_Offset        appdef.FieldName = "offset"
	Prp_BaseRecordID  appdef.FieldName = "baseRecordID"
	Prp_CBaseRecordID appdef.FieldName = "cBaseRecordID"
)
