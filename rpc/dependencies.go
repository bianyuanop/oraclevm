// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bianyuanop/oraclevm/genesis"
	"github.com/bianyuanop/oraclevm/oracle"
)

type Controller interface {
	Genesis() *genesis.Genesis
	Tracer() trace.Tracer
	GetTransaction(context.Context, ids.ID) (bool, int64, bool, uint64, error)
	GetBalanceFromState(context.Context, crypto.PublicKey) (uint64, error)
	GetHistoryFromState(uint64, uint64) ([]oracle.Entity, error)
	GetAvailableEntities() ([]*oracle.EntityCollectionMeta, error)
	GetEntitiesCollectionCount(entityIndex uint64) (uint64, error)
}
