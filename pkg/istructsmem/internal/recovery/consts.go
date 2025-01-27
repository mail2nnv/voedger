/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovery

import (
	"github.com/voedger/voedger/pkg/appdef"
)

// Partition recovery points view
var (
	prp_ViewName      appdef.QName     = appdef.NewQName(appdef.SysPackage, "prpView")
	prp_PID           appdef.FieldName = "pid"
	prp_WSID          appdef.FieldName = "wsid"
	prp_Offset        appdef.FieldName = "offset"
	prp_BaseRecordID  appdef.FieldName = "baseRecordID"
	prp_CBaseRecordID appdef.FieldName = "cBaseRecordID"
)
