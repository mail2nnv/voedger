/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package sequencer

import "github.com/voedger/voedger/pkg/istructs"

type changes struct {
	plogOffset istructs.Offset
	wlogOffset map[istructs.WSID]istructs.Offset
	recID      map[istructs.WSID]istructs.RecordID
	cRecID     map[istructs.WSID]istructs.RecordID
	count      int
}
