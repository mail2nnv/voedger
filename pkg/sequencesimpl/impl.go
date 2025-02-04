/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencesimpl

import (
	"context"
	"time"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
)

// # Supports:
//
//	sequences.ISequences
type Sequences struct {
	pid istructs.PartitionID
}

func new(pid istructs.PartitionID) *Sequences {
	return &Sequences{
		pid: pid,
	}
}

func (s *Sequences) Apply() {
	// TODO
}

func (s *Sequences) Discard() {
	// TODO
}

func (s *Sequences) Next(istructs.WSID, appdef.QName) uint64 {
	// TODO
	return 0
}

func (s *Sequences) Recovery(context.Context) {
	// TODO
}

func (s *Sequences) WaitForRecovery(time.Duration) bool {
	// TODO
	return false
}
