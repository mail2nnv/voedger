/*
 * Copyright (c) 2021-present unTill Pro, Ltd.
 */

package commandprocessor

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
)

var (
	ViewQNamePLogKnownOffsets = appdef.NewQName(appdef.SysPackage, "PLogKnownOffsets")
	ViewQNameWLogKnownOffsets = appdef.NewQName(appdef.SysPackage, "WLogKnownOffsets")
	opKindDesc                = map[appdef.OperationKind]string{
		appdef.OperationKind_Update:     "UPDATE",
		appdef.OperationKind_Insert:     "INSERT",
		appdef.OperationKind_Activate:   "ACTIVATE",
		appdef.OperationKind_Deactivate: "DEACTIVATE",
	}
)

const (
	// The renewal period of the partition recovery point (number of events).
	// Default is every 100 events.
	PartitionRecoveryPointRenewalPeriod istructs.Offset = 100
)
