/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovery

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/voedger/voedger/pkg/istorage"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem/internal/consts"
	"github.com/voedger/voedger/pkg/istructsmem/internal/utils"
	"github.com/voedger/voedger/pkg/istructsmem/internal/vers"
)

// # Supports:
//   - istructs.IRecovers
type Recovers struct {
	mutex   sync.RWMutex
	storage istorage.IAppStorage
	points  map[istructs.PartitionID]istructs.PartitionRecoveryPoint
}

func NewRecovers(storage istorage.IAppStorage) *Recovers {
	return &Recovers{
		mutex:   sync.RWMutex{},
		storage: storage,
		points:  make(map[istructs.PartitionID]istructs.PartitionRecoveryPoint),
	}
}

func (r *Recovers) Get(p istructs.PartitionID) istructs.PartitionRecoveryPoint {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if p, ok := r.points[p]; ok {
		return p
	}
	return istructs.PartitionRecoveryPoint{}
}

func (r *Recovers) Put(p istructs.PartitionID, point istructs.PartitionRecoveryPoint) (err error) {
	var pk, cc, data []byte
	pk = utils.ToBytes(consts.SysView_Recovers, ver01)
	cc = utils.ToBytes(uint64(p))
	if data, err = json.Marshal(point); err != nil {
		return err
	}

	if err = r.storage.Put(pk, cc, data); err != nil {
		return err
	}

	r.mutex.Lock()
	r.points[p] = point
	r.mutex.Unlock()

	return nil
}

// Prepare prepares the recovery points.
//
// Should once be called before other methods.
func (r *Recovers) Prepare(versions *vers.Versions) error {
	ver := versions.Get(vers.SysRecoversVersion)
	switch ver {
	case vers.UnknownVersion: // no sys.Recovers storage exists
		return nil
	case ver01:
		if err := r.load01(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown version of Recovery system view (%v): %w", ver, vers.ErrorInvalidVersion)
	}

	if ver != latestVersion {
		r.store()
	}

	return nil
}

func (r *Recovers) load01() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	pk := utils.ToBytes(consts.SysView_Recovers, ver01)
	return r.storage.Read(context.Background(), pk, nil, nil,
		func(ccols []byte, data []byte) (err error) {
			if len(ccols) != 8 {
				return fmt.Errorf("unexpected length of columns (%v) in Recovers system view: %w", len(ccols), io.ErrUnexpectedEOF)
			}

			pid := istructs.PartitionID(binary.BigEndian.Uint64(ccols))
			point := istructs.PartitionRecoveryPoint{}

			if err := json.Unmarshal(data, &point); err != nil {
				return fmt.Errorf("error unmarshalling PartitionRecoveryPoint: %w", err)
			}

			r.points[pid] = point

			return nil
		})
}

func (r *Recovers) store() {
}
