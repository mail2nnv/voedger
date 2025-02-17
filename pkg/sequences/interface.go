/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequences

import (
	"github.com/voedger/voedger/pkg/istructs"
)

// ISequencerFactory defines the interface for creating ISequencer instances
type ISequencerFactory interface {
	// New creates a new ISequencer instance for the given partition ID
	New(pid istructs.PartitionID) (ISequencer, cleanup func())
}

// NewFactory creates a new ISequencerFactory instance
func NewFactory(str istructs.IAppStructs) ISequencerFactory {
	return &sequencerFactory{str: str}
}
