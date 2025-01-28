/*
 * Copyright (c) 2025-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package recovery_test

import (
	"errors"
	"testing"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/appdef/builder"
	"github.com/voedger/voedger/pkg/goutils/testingu/require"
	"github.com/voedger/voedger/pkg/iratesce"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem"
	"github.com/voedger/voedger/pkg/istructsmem/internal/recovery"
	"github.com/voedger/voedger/pkg/istructsmem/internal/teststore"
	"github.com/voedger/voedger/pkg/istructsmem/internal/utils"
	payloads "github.com/voedger/voedger/pkg/itokens-payloads"
	"github.com/voedger/voedger/pkg/itokensjwt"
)

func TestRecovers(t *testing.T) {
	// test interface compatibility
	var _ istructs.IRecovers = &recovery.Recovers{}

	require := require.New(t)

	appName := istructs.AppQName_test1_app1
	appPartID := istructs.PartitionID(1)

	appConfigs := func() istructsmem.AppConfigsType {
		adb := builder.New()
		adb.AddPackage("test", "test.com/test")
		_ = adb.AddWorkspace(appdef.NewQName("test", "workspace"))
		cfgs := make(istructsmem.AppConfigsType, 1)
		cfg := cfgs.AddBuiltInAppConfig(appName, adb)
		cfg.SetNumAppWorkspaces(istructs.DefaultNumAppWorkspaces)
		return cfgs
	}()

	storage, storageProvider := teststore.New(appName)
	provider := istructsmem.Provide(
		appConfigs,
		iratesce.TestBucketsFactory,
		payloads.ProvideIAppTokensFactory(itokensjwt.TestTokensJWT()),
		storageProvider)

	app, err := provider.BuiltIn(appName)
	require.NoError(err)
	require.NotNil(app)

	recovers := app.Recovers()
	require.NotNil(recovers)

	prp, err := recovers.Get(appPartID)
	require.NoError(err)

	t.Run("Newly created IAppStructs should has empty PRP", func(t *testing.T) {
		require.Equal(appPartID, prp.PID())
		require.Equal(istructs.NullOffset, prp.PLogOffset())
		require.Empty(prp.Workspaces())
	})

	const (
		plogSize istructs.Offset = 1000
		wsCount                  = 10
	)

	testWS := make([]struct {
		wLog istructs.Offset
		id   istructs.RecordID
		cid  istructs.RecordID
	}, wsCount)

	for i := range testWS {
		testWS[i].wLog = istructs.FirstOffset
		testWS[i].id = istructs.NewRecordID(1)
		testWS[i].cid = istructs.NewRecordID(2)
	}

	t.Run("Should be ok to update PRP in cycle", func(t *testing.T) {
		for plog := istructs.FirstOffset; plog <= plogSize; plog++ {
			wsIdx := int(plog) % wsCount
			wsID := istructs.FirstBaseAppWSID + istructs.WSID(wsIdx)
			prp.Update(plog, wsID, testWS[wsIdx].wLog, testWS[wsIdx].id, testWS[wsIdx].cid)

			wsRP, ok := prp.Workspaces()[wsID]
			require.True(ok)
			require.NotNil(wsRP)
			require.Equal(wsID, wsRP.WSID())
			require.Equal(testWS[wsIdx].wLog, wsRP.WLogOffset())
			require.Equal(testWS[wsIdx].id, wsRP.BaseRecordID())
			require.Equal(testWS[wsIdx].cid, wsRP.CBaseRecordID())

			testWS[wsIdx].wLog++
			testWS[wsIdx].id += 2
			testWS[wsIdx].cid += 2
		}
	})

	t.Run("Should be ok to put PRP to storage", func(t *testing.T) {
		err := recovers.Put(prp)
		require.NoError(err)
	})

	t.Run("Should be ok to get PRP from storage", func(t *testing.T) {
		prp1, err := recovers.Get(appPartID)
		require.NoError(err)

		require.Equal(prp.PID(), prp1.PID())
		require.Equal(prp.PLogOffset(), prp1.PLogOffset())
		require.Equal(len(prp.Workspaces()), len(prp1.Workspaces()))
		for wsID, wsRP := range prp.Workspaces() {
			wsRP1, ok := prp1.Workspaces()[wsID]
			require.True(ok)
			require.NotNil(wsRP1)
			require.Equal(wsRP.WSID(), wsRP1.WSID())
			require.Equal(wsRP.WLogOffset(), wsRP1.WLogOffset())
			require.Equal(wsRP.BaseRecordID(), wsRP1.BaseRecordID())
			require.Equal(wsRP.CBaseRecordID(), wsRP1.CBaseRecordID())
		}
	})

	t.Run("Should be fail to get PRP if storage fails", func(t *testing.T) {
		testError := errors.New("test error")
		storage.ScheduleGetError(testError, nil, utils.ToBytes(istructs.FirstBaseAppWSID+5))
		_, err := recovers.Get(appPartID)
		require.Error(err, require.Is(testError))
	})
}
