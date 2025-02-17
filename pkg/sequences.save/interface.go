/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequences

import (
	"context"
	"time"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/istructs"
)

type ISequences interface {
	// Recovers sequence numbers from DB view and PLog.
	// Should be called from partition deployment
	Recovery(context.Context)

	// Waits while recovery finished.
	// Should be called by CP before PLog event building
	WaitForRecovery(time.Duration) bool

	// Returns next number in specified sequence for specified workspace.
	// Should be called by CP while PLog event building (CUDs, Arg, plog, wlog offset)
	Next(istructs.WSID, appdef.QName) uint64

	// Force to store changes in sequences numbers to DB view.
	// Should be called by CP after success PLog event writing
	Apply()

	// Force to discard all changes in sequences from last success Apply.
	// Should be called by CP after PLog event writing failed
	Discard()
}
