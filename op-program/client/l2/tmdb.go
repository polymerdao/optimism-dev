package l2

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	dbm "github.com/cosmos/iavl/db"
)

const HashSize = sha256.Size

var (
	_                         dbm.DB = &OracleIAVLKeyValueStore{}
	NodeKeyPrefix                    = []byte{'n'}
	ErrInvalidLegacyKeyPrefix        = fmt.Errorf("node keys must be prefixed with %x", NodeKeyPrefix)
)

// OracleIAVLKeyValueStore implements dbm.DB with all read access being routed through the underlying StateOracle
type OracleIAVLKeyValueStore struct {
	db     dbm.DB
	oracle StateOracle
	listen chan kvPair // used for debugging and test setup, consider removing later
}

// NewOracleBackedIAVLDB returns a new *OracleIAVLKeyValueStore
func NewOracleBackedIAVLDB(oracle StateOracle, listen chan kvPair) *OracleIAVLKeyValueStore {
	return &OracleIAVLKeyValueStore{
		db:     dbm.NewMemDB(),
		oracle: oracle,
		listen: listen,
	}
}

// Get satisfies db.DB
func (o OracleIAVLKeyValueStore) Get(key []byte) ([]byte, error) {
	has, err := o.db.Has(key)
	if err != nil {
		return nil, fmt.Errorf("checking in-memory db: %w", err)
	}
	if has {
		return o.db.Get(key)
	}
	if len(key) == HashSize+1 {
		if !bytes.HasPrefix(key, NodeKeyPrefix) {
			return nil, ErrInvalidLegacyKeyPrefix
		}
		key = bytes.TrimPrefix(key, NodeKeyPrefix)
	}
	if len(key) != HashSize {
		return nil, ErrInvalidKeyLength
	}
	return o.oracle.NodeByHash(*(*[HashSize]byte)(key)), nil
}

// Has satisfies db.DB
func (o OracleIAVLKeyValueStore) Has(key []byte) (bool, error) {
	return o.db.Has(key)
}

// Iterator satisfies db.DB
func (o OracleIAVLKeyValueStore) Iterator(start, end []byte) (dbm.Iterator, error) {
	return o.db.Iterator(start, end)
}

// ReverseIterator satisfies db.DB
func (o OracleIAVLKeyValueStore) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	return o.db.ReverseIterator(start, end)
}

// Close satisfies db.DB
func (o OracleIAVLKeyValueStore) Close() error {
	return nil
}

// NewBatch satisfies db.DB
func (o OracleIAVLKeyValueStore) NewBatch() dbm.Batch {
	return wrapBatch{
		batcher: o.db.NewBatch(),
		listen:  o.listen,
	}
}

// NewBatchWithSize satisfies db.DB
func (o OracleIAVLKeyValueStore) NewBatchWithSize(size int) dbm.Batch {
	return wrapBatch{
		batcher: o.db.NewBatchWithSize(size),
		listen:  o.listen,
	}
}

type kvPair struct {
	key   []byte
	value []byte
}

// used for debugging and test setup, consider removing later
type wrapBatch struct {
	batcher dbm.Batch
	listen  chan kvPair
}

var _ dbm.Batch = wrapBatch{}

// Set satisfies db.Batch
func (w wrapBatch) Set(key, val []byte) error {
	if w.listen != nil {
		w.listen <- kvPair{key, val}
	}
	return w.batcher.Set(key, val)
}

// Delete satisfies db.Batch
func (w wrapBatch) Delete(key []byte) error {
	return w.batcher.Delete(key)
}

// Write satisfies db.Batch
func (w wrapBatch) Write() error {
	return w.batcher.Write()
}

// WriteSync satisfies db.Batch
func (w wrapBatch) WriteSync() error {
	return w.batcher.WriteSync()
}

// Close satisfies db.Batch
func (w wrapBatch) Close() error {
	return w.batcher.Close()
}

// GetByteSize satisfies db.Batch
func (w wrapBatch) GetByteSize() (int, error) {
	return w.batcher.GetByteSize()
}
