/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package containers_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/voedger/voedger/pkg/appdef"
	"github.com/voedger/voedger/pkg/appdef/builder"
	"github.com/voedger/voedger/pkg/goutils/testingu"
	"github.com/voedger/voedger/pkg/istorage/mem"
	istorageimpl "github.com/voedger/voedger/pkg/istorage/provider"
	"github.com/voedger/voedger/pkg/istructs"
	"github.com/voedger/voedger/pkg/istructsmem/internal/consts"
	"github.com/voedger/voedger/pkg/istructsmem/internal/containers"
	"github.com/voedger/voedger/pkg/istructsmem/internal/teststore"
	"github.com/voedger/voedger/pkg/istructsmem/internal/utils"
	"github.com/voedger/voedger/pkg/istructsmem/internal/vers"
)

func TestNew(t *testing.T) {
	require := require.New(t)

	appName := istructs.AppQName_test1_app1

	sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
	storage, err := sp.AppStorage(appName)
	require.NoError(err)

	versions := vers.New()
	if err := versions.Prepare(storage); err != nil {
		panic(err)
	}

	containerName := "test"

	conts, err := containers.New(storage, versions,
		func() appdef.IAppDef {
			objName := appdef.NewQName("test", "object")
			adb := builder.New()
			adb.AddPackage("test", "test.com/test")
			wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
			wsb.AddObject(objName).
				AddContainer(containerName, objName, 0, 1)
			result, err := adb.Build()
			require.NoError(err)
			return result
		}())

	if err != nil {
		panic(err)
	}

	t.Run("basic Containers methods", func(t *testing.T) {

		check := func(conts *containers.Containers, name string) containers.ContainerID {
			id, err := conts.ID(name)
			require.NoError(err)
			require.NotEqual(containers.NullContainerID, id)

			n, err := conts.Name(id)
			require.NoError(err)
			require.Equal(name, n)

			return id
		}

		id := check(conts, containerName)

		t.Run("should be ok to load early stored names", func(t *testing.T) {
			versions1 := vers.New()
			if err := versions1.Prepare(storage); err != nil {
				panic(err)
			}

			containers1, err := containers.New(storage, versions, nil)
			if err != nil {
				panic(err)
			}

			require.Equal(id, check(containers1, containerName))
		})

		t.Run("should be ok to redeclare containers", func(t *testing.T) {
			versions2 := vers.New()
			if err := versions2.Prepare(storage); err != nil {
				panic(err)
			}

			containers2, err := containers.New(storage, versions,
				func() appdef.IAppDef {
					objName := appdef.NewQName("test", "object")
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					wsb.AddObject(objName).
						AddContainer(containerName, objName, 0, 1)
					result, err := adb.Build()
					require.NoError(err)
					return result
				}())
			if err != nil {
				panic(err)
			}

			require.Equal(id, check(containers2, containerName))
		})
	})

	t.Run("should be error if unknown container", func(t *testing.T) {
		id, err := conts.ID("unknown")
		require.Equal(containers.NullContainerID, id)
		require.ErrorIs(err, containers.ErrContainerNotFound)
	})

	t.Run("should be error if unknown id", func(t *testing.T) {
		n, err := conts.Name(containers.ContainerID(containers.MaxAvailableContainerID))
		require.Empty(n)
		require.ErrorIs(err, containers.ErrContainerIDNotFound)
	})
}

func TestErrors(t *testing.T) {
	require := require.New(t)

	appName := istructs.AppQName_test1_app1

	t.Run("should be error if unknown system view version", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysContainersVersion, containers.LatestVersion+1)

		_, err := containers.New(storage, versions, nil)
		require.ErrorIs(err, vers.ErrorInvalidVersion)
	})

	t.Run("should be error if invalid Container loaded from system view", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysContainersVersion, containers.LatestVersion)
		const badName = "-test-error-name-"
		storage.Put(utils.ToBytes(consts.SysView_Containers, containers.Ver01), []byte(badName), utils.ToBytes(containers.ContainerID(512)))

		_, err := containers.New(storage, versions, nil)
		require.ErrorIs(err, appdef.ErrInvalidError)
	})

	t.Run("should be ok if deleted Container loaded from system view", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysContainersVersion, containers.LatestVersion)
		storage.Put(utils.ToBytes(consts.SysView_Containers, containers.Ver01), []byte("deleted"), utils.ToBytes(containers.NullContainerID))

		_, err := containers.New(storage, versions, nil)
		require.NoError(err)
	})

	t.Run("should be error if invalid (small) ContainerID loaded from system view", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysContainersVersion, containers.LatestVersion)
		storage.Put(utils.ToBytes(consts.SysView_Containers, containers.Ver01), []byte("test"), utils.ToBytes(containers.ContainerID(1)))

		_, err := containers.New(storage, versions, nil)
		require.ErrorIs(err, containers.ErrWrongContainerID)
	})

	t.Run("should be error if too many Containers", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		_, err := containers.New(storage, versions,
			func() appdef.IAppDef {
				adb := builder.New()
				adb.AddPackage("test", "test.com/test")
				wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
				qName := appdef.NewQName("test", "test")
				obj := wsb.AddObject(qName)
				for i := 0; i <= containers.MaxAvailableContainerID; i++ {
					obj.AddContainer(fmt.Sprintf("cont_%d", i), qName, 0, 1)
				}
				result, err := adb.Build()
				require.NoError(err)
				return result
			}())
		require.ErrorIs(err, containers.ErrContainerIDsExceeds)
	})

	t.Run("should be error if write to storage failed", func(t *testing.T) {
		containerName := "testContainerName"
		writeError := errors.New("storage write error")

		t.Run("should be error if write some name failed", func(t *testing.T) {
			storage := teststore.NewStorage(appName)

			versions := vers.New()
			if err := versions.Prepare(storage); err != nil {
				panic(err)
			}

			storage.SchedulePutError(writeError, utils.ToBytes(consts.SysView_Containers, containers.Ver01), []byte(containerName))

			_, err := containers.New(storage, versions,
				func() appdef.IAppDef {
					objName := appdef.NewQName("test", "object")
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					wsb.AddObject(objName).
						AddContainer(containerName, objName, 0, 1)
					result, err := adb.Build()
					require.NoError(err)
					return result
				}())
			require.ErrorIs(err, writeError)
		})

		t.Run("should be error if write system view version failed", func(t *testing.T) {
			storage := teststore.NewStorage(appName)

			versions := vers.New()
			if err := versions.Prepare(storage); err != nil {
				panic(err)
			}

			storage.SchedulePutError(writeError, utils.ToBytes(consts.SysView_Versions), utils.ToBytes(vers.SysContainersVersion))

			_, err := containers.New(storage, versions,
				func() appdef.IAppDef {
					objName := appdef.NewQName("test", "object")
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					wsb.AddObject(objName).
						AddContainer(containerName, objName, 0, 1)
					result, err := adb.Build()
					require.NoError(err)
					return result
				}())
			require.ErrorIs(err, writeError)
		})
	})
}
