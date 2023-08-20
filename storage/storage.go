// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto"

	"github.com/bianyuanop/oraclevm/utils"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

// Metadata
// 0x0/ (tx)
//   -> [txID] => timestamp
//
// State
// / (height) => store in root
//   -> [heightPrefix] => height
// 0x0/ (balance)
//   -> [owner] => balance
// 0x1/ (hypersdk-incoming warp)
// 0x2/ (hypersdk-outgoing warp)

const (
	txPrefix = 0x0

	balancePrefix      = 0x0
	incomingWarpPrefix = 0x1
	outgoingWarpPrefix = 0x2

	// store entity uploading history
	entityPrefix = 0x3
	// store entity aggregation result
	entityAggregationResultPrefix = 0x4
	entityAggregationCachePrefix  = 0x5
)

var (
	failureByte = byte(0x0)
	successByte = byte(0x1)
	heightKey   = []byte{}

	// TODO: extend to other types
	balancePrefixPool = sync.Pool{
		New: func() any {
			return make([]byte, 1+crypto.PublicKeyLen)
		},
	}
)

// [txPrefix] + [txID]
func PrefixTxKey(id ids.ID) (k []byte) {
	// TODO: use packer?
	k = make([]byte, 1+consts.IDLen)
	k[0] = txPrefix
	copy(k[1:], id[:])
	return
}

func StoreTransaction(
	_ context.Context,
	db database.KeyValueWriter,
	id ids.ID,
	t int64,
	success bool,
	units uint64,
) error {
	k := PrefixTxKey(id)
	v := make([]byte, consts.Uint64Len+1+consts.Uint64Len)
	binary.BigEndian.PutUint64(v, uint64(t))
	if success {
		v[consts.Uint64Len] = successByte
	} else {
		v[consts.Uint64Len] = failureByte
	}
	binary.BigEndian.PutUint64(v[consts.Uint64Len+1:], units)
	return db.Put(k, v)
}

func GetTransaction(
	_ context.Context,
	db database.KeyValueReader,
	id ids.ID,
) (bool, int64, bool, uint64, error) {
	k := PrefixTxKey(id)
	v, err := db.Get(k)
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, false, 0, nil
	}
	if err != nil {
		return false, 0, false, 0, err
	}
	t := int64(binary.BigEndian.Uint64(v))
	success := true
	if v[consts.Uint64Len] == failureByte {
		success = false
	}
	units := binary.BigEndian.Uint64(v[consts.Uint64Len+1:])
	return true, t, success, units, nil
}

// [accountPrefix] + [address]
func PrefixBalanceKey(pk crypto.PublicKey) (k []byte) {
	k = balancePrefixPool.Get().([]byte)
	k[0] = balancePrefix
	copy(k[1:], pk[:])
	return
}

// If locked is 0, then account does not exist
func GetBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
) (uint64, error) {
	dbKey, bal, err := getBalance(ctx, db, pk)
	balancePrefixPool.Put(dbKey)
	return bal, err
}

func getBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
) ([]byte, uint64, error) {
	k := PrefixBalanceKey(pk)
	bal, err := innerGetBalance(db.GetValue(ctx, k))
	return k, bal, err
}

// Used to serve RPC queries
func GetBalanceFromState(
	ctx context.Context,
	f ReadState,
	pk crypto.PublicKey,
) (uint64, error) {
	k := PrefixBalanceKey(pk)
	values, errs := f(ctx, [][]byte{k})
	bal, err := innerGetBalance(values[0], errs[0])
	balancePrefixPool.Put(k)
	return bal, err
}

func innerGetBalance(
	v []byte,
	err error,
) (uint64, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func SetBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	balance uint64,
) error {
	k := PrefixBalanceKey(pk)
	return setBalance(ctx, db, k, balance)
}

func setBalance(
	ctx context.Context,
	db chain.Database,
	dbKey []byte,
	balance uint64,
) error {
	return db.Insert(ctx, dbKey, binary.BigEndian.AppendUint64(nil, balance))
}

func AddBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	amount uint64,
) error {
	dbKey, bal, err := getBalance(ctx, db, pk)
	if err != nil {
		return err
	}
	nbal, err := smath.Add64(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not add balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			utils.Address(pk),
			amount,
		)
	}
	return setBalance(ctx, db, dbKey, nbal)
}

func SubBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	amount uint64,
) error {
	dbKey, bal, err := getBalance(ctx, db, pk)
	if err != nil {
		return err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract balance (bal=%d, addr=%v, amount=%d)",
			ErrInvalidBalance,
			bal,
			utils.Address(pk),
			amount,
		)
	}
	if nbal == 0 {
		// If there is no balance left, we should delete the record instead of
		// setting it to 0.
		return db.Remove(ctx, dbKey)
	}
	return setBalance(ctx, db, dbKey, nbal)
}

func HeightKey() (k []byte) {
	return heightKey
}

func IncomingWarpKeyPrefix(sourceChainID ids.ID, msgID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen*2)
	k[0] = incomingWarpPrefix
	copy(k[1:], sourceChainID[:])
	copy(k[1+consts.IDLen:], msgID[:])
	return k
}

func OutgoingWarpKeyPrefix(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = outgoingWarpPrefix
	copy(k[1:], txID[:])
	return k
}

// [entityPrefix] + [txID] + [entityIndex]
func PrefixEntityKey(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = entityPrefix
	copy(k[1:], txID[:])

	return
}

func PackEntity(entityIndex uint64, entityType uint64, tick int64, publisher crypto.PublicKey, payload []byte) (v []byte) {
	v = make([]byte, consts.Uint64Len*3+crypto.PublicKeyLen+len(payload))

	binary.BigEndian.PutUint64(v, entityIndex)
	binary.BigEndian.PutUint64(v[consts.Uint64Len:], entityType)
	binary.BigEndian.PutUint64(v[consts.Uint64Len*2:], uint64(tick))
	copy(v[consts.Uint64Len*3:], publisher[:])
	copy(v[consts.Uint64Len*3+crypto.PublicKeyLen:], payload)

	return
}

func UnpackEntity(v []byte) (
	entityIndex uint64,
	entityType uint64,
	tick int64,
	publisher crypto.PublicKey,
	payload []byte,
) {
	entityIndex = binary.BigEndian.Uint64(v)
	entityType = binary.BigEndian.Uint64(v[consts.Uint64Len:])
	tick = int64(binary.BigEndian.Uint64(v[consts.Uint64Len*2:]))
	publisher = crypto.PublicKey(v[consts.Uint64Len*3:])

	payload = make([]byte, len(v)-consts.Uint64Len*3+crypto.PublicKeyLen)

	copy(payload, v[consts.Uint64Len*3+crypto.PublicKeyLen:])

	return
}

func StoreEntity(
	ctx context.Context,
	db chain.Database,
	txID ids.ID,
	entityType uint64,
	entityIndex uint64,
	tick int64,
	publisher crypto.PublicKey,
	payload []byte,
) error {
	k := PrefixEntityKey(txID)
	v := PackEntity(entityIndex, entityType, tick, publisher, payload)

	return db.Insert(ctx, k, v)
}

func GetEntity(
	ctx context.Context,
	db chain.Database,
	txID ids.ID,
) (
	exists bool,
	tick int64,
	entityIndex uint64,
	entityType uint64,
	publisher crypto.PublicKey,
	payload []byte,
	e error,
) {
	k := PrefixEntityKey(txID)

	v, err := db.GetValue(ctx, k)

	if errors.Is(err, database.ErrNotFound) {
		return false, 0, 0, 0, crypto.EmptyPublicKey, make([]byte, 0), nil
	}

	if err != nil {
		return false, 0, 0, 0, crypto.EmptyPublicKey, make([]byte, 0), err
	}

	entityIndex, entityType, tick, publisher, payload = UnpackEntity(v)

	return
}

func PrefixAggregationResult(tick int64, entityIndex uint64) (k []byte) {
	k = make([]byte, 1+consts.Uint64Len*2)
	k[0] = entityAggregationResultPrefix
	binary.BigEndian.PutUint64(k[1:], uint64(tick))
	binary.BigEndian.PutUint64(k[1+consts.Uint64Len:], entityIndex)

	return
}

func StoreAggregationResult(
	ctx context.Context,
	db database.KeyValueWriter,
	entityType uint64,
	entityIndex uint64,
	tick int64,
	payload []byte,
) error {
	k := PrefixAggregationResult(tick, entityIndex)

	v := PackEntity(entityIndex, entityType, tick, crypto.EmptyPublicKey, payload)

	return db.Put(k, v)
}

func GetAggregationResult(
	ctx context.Context,
	db database.KeyValueReader,
	entityIndex uint64,
	tick int64, // same with block timestamp
) (entityType uint64, payload []byte, e error) {
	k := PrefixAggregationResult(tick, entityIndex)

	v, err := db.Get(k)

	if err != nil {
		return 0, make([]byte, 0), err
	}

	_, entityType, _, _, payload = UnpackEntity(v)

	return
}

func PrefixAggregationCacheResult(entityIndex uint64) (k []byte) {
	k = make([]byte, 1+consts.Uint64Len)
	k[0] = entityAggregationCachePrefix
	binary.BigEndian.PutUint64(k[1:], entityIndex)

	return
}

func CacheAggregationResult(
	ctx context.Context,
	db database.KeyValueWriter,
	entityType uint64,
	entityIndex uint64,
	tick int64,
	payload []byte,
) error {
	k := PrefixAggregationCacheResult(entityIndex)
	v := PackEntity(entityIndex, entityType, tick, crypto.EmptyPublicKey, payload)

	return db.Put(k, v)
}

func GetCachedAggregationResult(
	ctx context.Context,
	db chain.Database,
	entityIndex uint64,
) (entityType uint64, payload []byte, e error) {
	k := PrefixAggregationCacheResult(entityIndex)

	v, err := db.GetValue(ctx, k)
	if err != nil {
		return 0, make([]byte, 0), err
	}

	_, entityType, _, _, payload = UnpackEntity(v)

	return
}
