/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequences

import (
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/sequences/sequencer"
)

type sequencerFactory struct {
	str istructs.IAppStructs
}

func (f *sequencerFactory) New(pid istructs.PartitionID) (ISequencer, cleanup func()) {
	return sequencer.NewSequencer(pid, f.str)
}
