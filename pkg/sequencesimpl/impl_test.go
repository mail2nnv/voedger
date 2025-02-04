/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencesimpl_test

import (
	"testing"

	"github.com/voedger/voedger/pkg/sequences"
	"github.com/voedger/voedger/pkg/sequencesimpl"
)

func TestNew(t *testing.T) {
	var _ sequences.ISequences = sequencesimpl.New(nil, 0)
	t.Skip("TODO")
}
