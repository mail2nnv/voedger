/*
 * Copyright (c) 2020-present unTill Pro, Ltd.
 * @author Denis Gribanov
 */

package builtin

import (
	"github.com/voedger/voedger/pkg/appdef"
)

var (
	// Deprecated: use c.sys.CUD instead. Kept to not to break existing events only
	QNameCommandInit              = appdef.NewQName(appdef.SysPackage, "Init")
	QNameViewRecordsRegistry      = appdef.NewQName(appdef.SysPackage, "RecordsRegistry")
	qNameRecordsRegistryProjector = appdef.NewQName(appdef.SysPackage, "RecordsRegistryProjector")
)

const (
	field_ExistingQName = "ExistingQName"
	field_NewQName      = "NewQName"
	MaxCUDs             = 100
	Field_IDHi          = "IDHi"
	Field_ID            = "ID"
	Field_WLogOffset    = "WLogOffset"
	field_QName         = "QName"
	field_IsActive      = "IsActive"
	registryViewBits    = 18
)
