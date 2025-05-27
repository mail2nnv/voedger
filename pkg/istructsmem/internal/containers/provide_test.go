/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package containers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/appdef/builder"
	"github.com/voedger/voedger/pkg/goutils/testingu"
	"github.com/voedger/voedger/pkg/istorage/mem"
	istorageimpl "github.com/voedger/voedger/pkg/istorage/provider"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem/internal/containers"
	"github.com/voedger/voedger/pkg/istructsmem/internal/vers"
)

func Test_BasicUsage(t *testing.T) {
	sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
	storage, _ := sp.AppStorage(istructs.AppQName_test1_app1)

	versions := vers.New()
	if err := versions.Prepare(storage); err != nil {
		panic(err)
	}

	testName := "test"
	adb := builder.New()
	adb.AddPackage("test", "test.com/test")
	wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))

	wsb.AddObject(appdef.NewQName("test", "obj")).
		AddContainer(testName, appdef.NewQName("test", "obj"), 0, appdef.Occurs_Unbounded)
	appDef, err := adb.Build()
	if err != nil {
		panic(err)
	}

	conts, err := containers.New(storage, versions, appDef)
	if err != nil {
		panic(err)
	}

	require := require.New(t)
	t.Run("basic Containers methods", func(t *testing.T) {
		id, err := conts.ID(testName)
		require.NoError(err)
		require.NotEqual(containers.NullContainerID, id)

		n, err := conts.Name(id)
		require.NoError(err)
		require.Equal(testName, n)

		t.Run("load early stored names", func(t *testing.T) {
			otherVersions := vers.New()
			if err := otherVersions.Prepare(storage); err != nil {
				panic(err)
			}

			otherConts, err := containers.New(storage, versions, nil)
			if err != nil {
				panic(err)
			}

			id1, err := otherConts.ID(testName)
			require.NoError(err)
			require.Equal(id, id1)

			n1, err := otherConts.Name(id)
			require.NoError(err)
			require.Equal(testName, n1)
		})
	})
}
