/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package containers

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istorage"
	"github.com/voedger/voedger/pkg/istructsmem/internal/vers"
)

// Constructs and return new containers system view
func New(storage istorage.IAppStorage, versions *vers.Versions, appDef appdef.IAppDef) (*Containers, error) {
	c := new()
	if err := c.prepare(storage, versions, appDef); err != nil {
		return nil, err
	}
	return c, nil
}
