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
	vr      istructs.IViewRecords
	pid     istructs.PartitionID
	last    map[istructs.WSID]map[appdef.QName]uint64
	changes *map[istructs.WSID]map[appdef.QName]uint64
	view    *view
}

func newSequences(vr istructs.IViewRecords, pid istructs.PartitionID) *Sequences {
	return &Sequences{
		vr:   vr,
		pid:  pid,
		last: make(map[istructs.WSID]map[appdef.QName]uint64),
		view: newView(vr, pid),
	}
}

func (s *Sequences) Apply() {
	if changes := s.changes; changes != nil {
		s.changes = nil
		s.view.apply(changes)
		for ws, ss := range *changes {
			if _, ok := s.last[ws]; !ok {
				s.last[ws] = make(map[appdef.QName]uint64)
			}
			for seq, last := range ss {
				s.last[ws][seq] = last
			}
		}
	}
}

func (s *Sequences) Discard() {
	s.changes = nil
}

func (s *Sequences) Next(ws istructs.WSID, seq appdef.QName) uint64 {
	if s.changes == nil {
		s.changes = new(map[istructs.WSID]map[appdef.QName]uint64)
	}

	var (
		ok   bool
		ss   map[appdef.QName]uint64
		last uint64
	)

	if ss, ok = (*s.changes)[ws]; !ok {
		ss = make(map[appdef.QName]uint64)
		(*s.changes)[ws] = ss
	}

	if last, ok = ss[seq]; !ok {
		last = s.last[ws][seq]
	}

	ss[seq] = last + 1

	return last
}

func (s *Sequences) Recovery(context.Context) {
	// TODO
}

func (s *Sequences) WaitForRecovery(time.Duration) bool {
	// TODO
	return false
}
