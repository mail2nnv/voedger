/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package qnames

import (
	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istorage"
	"github.com/voedger/voedger/pkg/istructsmem/internal/vers"
)

// Create and return new QNames
func New(storage istorage.IAppStorage, versions *vers.Versions, appDef appdef.IAppDef) (*QNames, error) {
	q := newQNames()
	if err := q.prepare(storage, versions, appDef); err != nil {
		return nil, err
	}
	return q, nil
}

// Renames QName from old to new. QNameID previously used by old will be used by new.
func Rename(storage istorage.IAppStorage, oldQName, newQName appdef.QName) error {
	return renameQName(storage, oldQName, newQName)
}
