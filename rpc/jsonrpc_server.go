// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/bianyuanop/oraclevm/genesis"
	"github.com/bianyuanop/oraclevm/oracle"
	"github.com/bianyuanop/oraclevm/utils"
)

type JSONRPCServer struct {
	c Controller
}

func NewJSONRPCServer(c Controller) *JSONRPCServer {
	return &JSONRPCServer{c}
}

type GenesisReply struct {
	Genesis *genesis.Genesis `json:"genesis"`
}

func (j *JSONRPCServer) Genesis(_ *http.Request, _ *struct{}, reply *GenesisReply) (err error) {
	reply.Genesis = j.c.Genesis()
	return nil
}

type TxArgs struct {
	TxID ids.ID `json:"txId"`
}

type TxReply struct {
	Timestamp int64  `json:"timestamp"`
	Success   bool   `json:"success"`
	Units     uint64 `json:"units"`
}

func (j *JSONRPCServer) Tx(req *http.Request, args *TxArgs, reply *TxReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Tx")
	defer span.End()

	found, t, success, units, err := j.c.GetTransaction(ctx, args.TxID)
	if err != nil {
		return err
	}
	if !found {
		return ErrTxNotFound
	}
	reply.Timestamp = t
	reply.Success = success
	reply.Units = units
	return nil
}

type BalanceArgs struct {
	Address string `json:"address"`
}

type BalanceReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()

	addr, err := utils.ParseAddress(args.Address)
	if err != nil {
		return err
	}
	balance, err := j.c.GetBalanceFromState(ctx, addr)
	if err != nil {
		return err
	}
	reply.Amount = balance
	return err
}

type HistoryArgs struct {
	EntityIndex uint64 `json:"index"`
	Limit       uint64 `json:"limit"`
}

type HistoryReply struct {
	History [][]byte `json:"history"`
	Length  int      `json:"length"`
}

func (j *JSONRPCServer) History(req *http.Request, args *HistoryArgs, reply *HistoryReply) error {
	_, span := j.c.Tracer().Start(req.Context(), "Server.History")
	defer span.End()

	history, err := j.c.GetHistoryFromState(args.EntityIndex, args.Limit)
	if err != nil {
		return err
	}

	reply.Length = len(history)
	reply.History = make([][]byte, reply.Length)
	for i := 0; i < int(args.Limit); i++ {
		reply.History[i] = history[i].Marshal()
	}

	return nil
}

type EntitiesMetaArgs struct{}

type EntitiesMetaReply struct {
	Metas []*oracle.EntityCollectionMeta `json:"metas"`
}

func (j *JSONRPCServer) Entities(req *http.Request, _ *EntitiesMetaArgs, reply *EntitiesMetaReply) error {
	_, span := j.c.Tracer().Start(req.Context(), "Server.Entities")
	defer span.End()

	metas, err := j.c.GetAvailableEntities()
	// never happen
	if err != nil {
		return err
	}

	reply.Metas = metas

	return nil
}

type EntityCollectionCountArgs struct {
	EntityIndex uint64 `json:"index"`
}

type EntityCollectionCountReply struct {
	Count uint64 `json:"count"`
}

// Entity Collection Count
func (j *JSONRPCServer) Eccount(req *http.Request, args *EntityCollectionCountArgs, reply *EntityCollectionCountReply) error {
	_, span := j.c.Tracer().Start(req.Context(), "Server.EntityCollectionCount")
	defer span.End()

	count, err := j.c.GetEntitiesCollectionCount(args.EntityIndex)
	// never happen
	if err != nil {
		return err
	}

	reply.Count = count

	return nil
}
