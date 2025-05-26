/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package rows

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem/internal/containers"
	"github.com/voedger/voedger/pkg/istructsmem/internal/dynobuf"
)

// Application configuration interface used row types.
type AppConfig interface {
	AppDef() appdef.IAppDef
	AppTokens() istructs.IAppTokens
	DynoBufSchemes() *dynobuf.DynoBufSchemes
	ContainerID(name string) (containers.ContainerID, error)
	QNameID(appdef.QName) (istructs.QNameID, error)
}
