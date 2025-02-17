/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequences

import (
	"time"

	"github.com/voedger/voedger/pkg/istructs"
)

// ISequencer provides methods to manage log offsets and record IDs
type ISequencer interface {
	// StartEvent reserves offsets (plog and wlog) with a wait timeout
	StartEvent(wait time.Duration, ws istructs.WSID) (plogOffset, wlogOffset istructs.Offset)

	// NextRecID returns the next available record ID for ODoc/ORecord/WDoc/WRecord
	NextRecID() istructs.RecordID

	// NextCRecID returns the next available record ID for CDoc/CRecord
	NextCRecID() istructs.RecordID

	// FinishEvent finalizes the current event
	FinishEvent()

	// CancelEvent aborts the ongoing event
	CancelEvent()
}
