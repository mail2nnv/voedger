/*
 * Copyright (c) 2021-present Sigma-Soft, Ltd.
 * @author: Nikolay Nikitin
 */

package qnames_test

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
	"github.com/voedger/voedger/pkg/istructsmem/internal/qnames"
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

	defName := appdef.NewQName("test", "doc")

	names, err := qnames.New(storage, versions,
		func() appdef.IAppDef {
			adb := builder.New()
			adb.AddPackage("test", "test.com/test")
			ws := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
			ws.AddCDoc(defName)
			appDef, err := adb.Build()
			require.NoError(err)
			return appDef
		}())
	if err != nil {
		panic(err)
	}

	t.Run("basic QNames methods", func(t *testing.T) {

		check := func(names *qnames.QNames, name appdef.QName) istructs.QNameID {
			id, err := names.ID(name)
			require.NoError(err)
			require.NotEqual(istructs.NullQNameID, id)

			n, err := names.QName(id)
			require.NoError(err)
			require.Equal(name, n)

			return id
		}

		sID := check(names, defName)

		t.Run("should be ok to load early stored names", func(t *testing.T) {
			versions1 := vers.New()
			if err := versions1.Prepare(storage); err != nil {
				panic(err)
			}

			names1, err := qnames.New(storage, versions, nil)
			if err != nil {
				panic(err)
			}

			require.Equal(sID, check(names1, defName))
		})

		t.Run("should be ok to redeclare names", func(t *testing.T) {
			versions2 := vers.New()
			if err := versions2.Prepare(storage); err != nil {
				panic(err)
			}

			names2, err := qnames.New(storage, versions,
				func() appdef.IAppDef {
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					ws := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					ws.AddCDoc(defName)
					appDef, err := adb.Build()
					require.NoError(err)
					return appDef
				}())
			if err != nil {
				panic(err)
			}

			require.Equal(sID, check(names2, defName))
		})
	})

	t.Run("should be error if unknown name", func(t *testing.T) {
		id, err := names.ID(appdef.NewQName("test", "unknown"))
		require.Equal(istructs.NullQNameID, id)
		require.ErrorIs(err, qnames.ErrNameNotFound)
	})

	t.Run("should be error if unknown id", func(t *testing.T) {
		n, err := names.QName(istructs.QNameID(qnames.MaxAvailableQNameID))
		require.Equal(appdef.NullQName, n)
		require.ErrorIs(err, qnames.ErrIDNotFound)
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

		versions.Put(vers.SysQNamesVersion, qnames.LatestVersion+1)

		_, err := qnames.New(storage, versions, nil)
		require.ErrorIs(err, vers.ErrorInvalidVersion)
	})

	t.Run("should be error if invalid QName loaded from system view ", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysQNamesVersion, qnames.LatestVersion)
		const badName = "-test.error.qname-"
		storage.Put(utils.ToBytes(consts.SysView_QNames, qnames.Ver01), []byte(badName), utils.ToBytes(istructs.QNameID(512)))

		_, err := qnames.New(storage, versions, nil)
		require.ErrorIs(err, appdef.ErrConvertError)
		require.ErrorContains(err, badName)
	})

	t.Run("should be ok if deleted QName loaded from system view ", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysQNamesVersion, qnames.LatestVersion)
		storage.Put(utils.ToBytes(consts.SysView_QNames, qnames.Ver01), []byte("test.deleted"), utils.ToBytes(istructs.NullQNameID))

		_, err := qnames.New(storage, versions, nil)
		require.NoError(err)
	})

	t.Run("should be error if invalid (small) istructs.QNameID loaded from system view ", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		versions.Put(vers.SysQNamesVersion, qnames.LatestVersion)
		storage.Put(utils.ToBytes(consts.SysView_QNames, qnames.Ver01), []byte(istructs.QNameForError.String()), utils.ToBytes(istructs.QNameIDForError))

		_, err := qnames.New(storage, versions, nil)
		require.ErrorIs(err, qnames.ErrWrongQNameID)
		require.ErrorContains(err, fmt.Sprintf("unexpected ID (%v)", istructs.QNameIDForError))
	})

	t.Run("should be error if too many QNames", func(t *testing.T) {
		sp := istorageimpl.Provide(mem.Provide(testingu.MockTime))
		storage, _ := sp.AppStorage(appName)

		versions := vers.New()
		if err := versions.Prepare(storage); err != nil {
			panic(err)
		}

		_, err := qnames.New(storage, versions,
			func() appdef.IAppDef {
				adb := builder.New()
				adb.AddPackage("test", "test.com/test")
				wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
				for i := 0; i <= qnames.MaxAvailableQNameID; i++ {
					wsb.AddObject(appdef.NewQName("test", fmt.Sprintf("name_%d", i)))
				}
				appDef, err := adb.Build()
				require.NoError(err)
				return appDef
			}())
		require.ErrorIs(err, qnames.ErrQNameIDsExceeds)
	})

	t.Run("should be error if write to storage failed", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}
		qName := appdef.NewQName("test", "test")
		writeError := errors.New("storage write error")

		t.Run("should be error if write some name failed", func(t *testing.T) {
			storage := teststore.NewStorage(appName)

			versions := vers.New()
			if err := versions.Prepare(storage); err != nil {
				panic(err)
			}

			storage.SchedulePutError(writeError, utils.ToBytes(consts.SysView_QNames, qnames.Ver01), []byte(qName.String()))

			_, err := qnames.New(storage, versions,
				func() appdef.IAppDef {
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					wsb.AddObject(qName)
					appDef, err := adb.Build()
					require.NoError(err)
					return appDef
				}())
			require.ErrorIs(err, writeError)
		})

		t.Run("should be error if write system view version failed", func(t *testing.T) {
			storage := teststore.NewStorage(appName)

			versions := vers.New()
			if err := versions.Prepare(storage); err != nil {
				panic(err)
			}

			storage.SchedulePutError(writeError, utils.ToBytes(consts.SysView_Versions), utils.ToBytes(vers.SysQNamesVersion))

			_, err := qnames.New(storage, versions,
				func() appdef.IAppDef {
					adb := builder.New()
					adb.AddPackage("test", "test.com/test")
					wsb := adb.AddWorkspace(appdef.NewQName("test", "workspace"))
					wsb.AddObject(qName)
					appDef, err := adb.Build()
					require.NoError(err)
					return appDef
				}())
			require.ErrorIs(err, writeError)
		})
	})
}
