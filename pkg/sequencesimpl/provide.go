/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencesimpl

import "github.com/voedger/voedger/pkg/istructs"

func New(pid istructs.PartitionID) *Sequences {
	return new(pid)
}
