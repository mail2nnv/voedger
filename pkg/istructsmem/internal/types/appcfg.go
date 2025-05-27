/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package types

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem/internal/containers"
	"github.com/voedger/voedger/pkg/istructsmem/internal/dynobuf"
	"github.com/voedger/voedger/pkg/istructsmem/internal/qnames"
)

// Application configuration interface used row types.
type AppConfig interface {
	AppDef() appdef.IAppDef
	AppTokens() istructs.IAppTokens
	DynoBufSchemes() *dynobuf.DynoBufSchemes
	Containers() *containers.Containers
	QNames() *qnames.QNames
}
