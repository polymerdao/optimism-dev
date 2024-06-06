package l2

import (
	"testing"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/iavl"

	"github.com/ethereum-optimism/optimism/op-program/client/l2/test"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var (
	tKey1 = []byte("k1")
	tVal1 = []byte("v1")

	tKey2 = []byte("k2")
	tVal2 = []byte("v2")

	tKey3 = []byte("k3")
	tVal3 = []byte("v3")
)

// Should implement the KeyValueStore API
var _ dbm.DB = (*OracleIAVLKeyValueStore)(nil)

func TestDBGet(t *testing.T) {
	t.Run("MissingKeyPrefix", func(t *testing.T) {
		key := common.HexToHash("0xAA4488")
		keyBytes := append([]byte{'a'}, key.Bytes()...)
		oracle := test.NewStubStateOracle(t)
		iavlDB := NewOracleBackedIAVLDB(oracle, nil)
		val, err := iavlDB.Get(keyBytes)
		require.ErrorIs(t, err, ErrInvalidLegacyKeyPrefix)
		require.Nil(t, val)
	})

	t.Run("IncorrectLengthKey", func(t *testing.T) {
		key := []byte{1, 2, 3}
		oracle := test.NewStubStateOracle(t)
		iavlDB := NewOracleBackedIAVLDB(oracle, nil)
		val, err := iavlDB.Get(key)
		require.ErrorIs(t, err, ErrInvalidKeyLength)
		require.Nil(t, val)
	})

	t.Run("KnownKeyNoPrefix", func(t *testing.T) {
		keyHash := common.HexToHash("0xAA4488")
		expected := []byte{2, 6, 3, 8}
		oracle := test.NewStubStateOracle(t)
		oracle.Data[keyHash] = expected
		iavlDB := NewOracleBackedIAVLDB(oracle, nil)
		val, err := iavlDB.Get(keyHash.Bytes())
		require.NoError(t, err)
		require.Equal(t, expected, val)
	})

	t.Run("KnownKeyCorrectPrefix", func(t *testing.T) {
		keyHash := common.HexToHash("0xAA4488")
		keyBytes := append(NodeKeyPrefix, keyHash.Bytes()...)
		expected := []byte{2, 6, 3, 8}
		oracle := test.NewStubStateOracle(t)
		oracle.Data[keyHash] = expected
		iavlDB := NewOracleBackedIAVLDB(oracle, nil)
		val, err := iavlDB.Get(keyBytes)
		require.NoError(t, err)
		require.Equal(t, expected, val)
	})
}

func TestSupportsIAVLOperations(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		oracle := test.NewStubStateOracle(t)
		listener := make(chan kvPair)
		mutableIAVL := setupOracleBackedMutableTree(t, oracle, listener, nil)
		testGet := func(exists bool) {
			v, err := mutableIAVL.Get(tKey1)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal1, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
			v, err = mutableIAVL.Get(tKey2)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal2, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
			v, err = mutableIAVL.Get(tKey3)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal3, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
		}

		testGet(false)

		// collect all the low-level dbm.DB insert values
		// this is done because it is difficult to reason about ahead of time what the actual nodes and their keys
		// will look like in the tree (not the application layer keys, which are what we insert at the IAVl level below)
		kvPairs := make([]kvPair, 0)
		go func() {
			for {
				select {
				case kvPair := <-listener:
					kvPairs = append(kvPairs, kvPair)
				}
			}
		}()

		updated, err := mutableIAVL.Set(tKey1, tVal1)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		updated, err = mutableIAVL.Set(tKey2, tVal2)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		updated, err = mutableIAVL.Set(tKey3, tVal3)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		testGet(true)

		// Save to tree.ImmutableTree
		_, version, err := mutableIAVL.SaveVersion()
		require.NoError(t, err)
		require.Equal(t, int64(1), version)

		time.Sleep(2 * time.Second) // wait to collect all nodes off listen channel
		testGet(true)
		rootHash, err := mutableIAVL.Hash()
		require.NoError(t, err)

		v, ok, err := mutableIAVL.Remove(tKey1)
		require.NoError(t, err)
		require.True(t, ok, "key should be removed")
		require.Equal(t, tVal1, v, "key should exist")

		v, err = mutableIAVL.Get(tKey1)
		require.NoError(t, err)
		require.Nil(t, v, "key should not exist")

		// Now create a new mutable tree with an oracle that has been pre-populated with the nodes we saw above
		oracle = test.NewStubStateOracle(t)
		for _, kvPair := range kvPairs {
			oracle.Data[common.BytesToHash(kvPair.key)] = kvPair.value
		}
		mutableIAVL = setupOracleBackedMutableTree(t, oracle, nil, rootHash)

		// And we should be able to fetch all the application level keys without having to insert them first
		testGet(true)
	})

	t.Run("Iterator", func(t *testing.T) {
		oracle := test.NewStubStateOracle(t)
		listener := make(chan kvPair)
		mutableIAVL := setupOracleBackedMutableTree(t, oracle, listener, nil)
		testGet := func(exists bool) {
			v, err := mutableIAVL.Get(tKey1)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal1, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
			v, err = mutableIAVL.Get(tKey2)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal2, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
			v, err = mutableIAVL.Get(tKey3)
			require.NoError(t, err)
			if exists {
				require.Equal(t, tVal3, v, "key should exist")
			} else {
				require.Nil(t, v, "key should not exist")
			}
		}

		testGet(false)

		// collect all the low-level dbm.DB insert values
		// this is done because it is difficult to reason about ahead of time what the actual nodes and their keys
		// will look like in the tree (not the application layer keys, which are what we insert at the IAVl level below)
		kvPairs := make([]kvPair, 0)
		go func() {
			for {
				select {
				case kvPair := <-listener:
					kvPairs = append(kvPairs, kvPair)
				}
			}
		}()

		updated, err := mutableIAVL.Set(tKey1, tVal1)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		updated, err = mutableIAVL.Set(tKey2, tVal2)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		updated, err = mutableIAVL.Set(tKey3, tVal3)
		require.NoError(t, err)
		require.Equal(t, updated, false, "new key set: nothing to update")

		testGet(true)

		// Save to tree.ImmutableTree
		_, version, err := mutableIAVL.SaveVersion()
		require.NoError(t, err)
		require.Equal(t, int64(1), version)

		time.Sleep(2 * time.Second) // wait to collect all nodes off listen channel
		testGet(true)
		rootHash, err := mutableIAVL.Hash()
		require.NoError(t, err)

		v, ok, err := mutableIAVL.Remove(tKey1)
		require.NoError(t, err)
		require.True(t, ok, "key should be removed")
		require.Equal(t, tVal1, v, "key should exist")

		v, err = mutableIAVL.Get(tKey1)
		require.NoError(t, err)
		require.Nil(t, v, "key should not exist")

		// Now create a new mutable tree with an oracle that has been pre-populated with the nodes we saw above
		oracle = test.NewStubStateOracle(t)
		for _, kvPair := range kvPairs {
			oracle.Data[common.BytesToHash(kvPair.key)] = kvPair.value
		}
		mutableIAVL = setupOracleBackedMutableTree(t, oracle, nil, rootHash)

		// And we should be able to fetch all the application level keys without having to insert them first
		testGet(true)

		iterator, err := mutableIAVL.Iterator([]byte{0x6b, 0x31}, []byte{0x6c, 0x33}, true)
		require.NoError(t, err)
		expectedKeys := [][]byte{tKey1, tKey2, tKey3}
		expectedValues := [][]byte{tVal1, tVal2, tVal3}
		collectedKeys := make([][]byte, 0, 3)
		collectedValues := make([][]byte, 0, 3)
		for {
			collectedKeys = append(collectedKeys, iterator.Key())
			collectedValues = append(collectedValues, iterator.Value())
			iterator.Next()
			if !iterator.Valid() {
				break
			}
		}
		require.Equal(t, expectedKeys, collectedKeys)
		require.Equal(t, expectedValues, collectedValues)
	})
}

func setupOracleBackedMutableTree(t *testing.T, oracle *test.StubStateOracle, listen chan kvPair, rootHash []byte) *iavl.MutableTree {
	iavlDB := NewOracleBackedIAVLDB(oracle, listen)
	mut, err := iavl.NewMutableTree(iavlDB, 0, false, rootHash, true)
	require.NoError(t, err)
	return mut
}
