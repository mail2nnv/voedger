/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencesimpl

import "github.com/voedger/voedger/pkg/istructs"

func New(vr istructs.IViewRecords, pid istructs.PartitionID) *Sequences {
	return newSequences(vr, pid)
}
