// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/bianyuanop/oraclevm/actions"
	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use: "action",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get balance info
		balance, err := handler.GetBalance(ctx, bcli, priv.PublicKey())
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", ids.Empty, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.Transfer{
			To:    recipient,
			Value: amount,
		}, cli, bcli, factory, true)
		return err
	},
}

var uploadCmd = &cobra.Command{
	Use: "upload_entity",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		entityIndex, err := handler.Root().PromptChoice("index", 1)
		if err != nil {
			return err
		}

		entityType, err := handler.Root().PromptChoice("type", 1)
		if err != nil {
			return err
		}

		payload, err := handler.Root().PromptString("payload", 0, 500)

		if err != nil {
			return err
		}

		_, _, err = sendAndWait(ctx, nil, &actions.UploadEntity{
			EntityIndex: uint64(entityIndex),
			EntityType:  uint64(entityType),
			Payload:     []byte(payload),
		}, cli, bcli, factory, true)

		return err
	},
}

var queryCmd = &cobra.Command{
	Use: "query",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		entityIndex, err := handler.Root().PromptChoice("index", 1)
		if err != nil {
			return err
		}

		warpQuery := actions.WarpQuery{
			EntityIndex:        uint64(entityIndex),
			DestinationChainID: ids.GenerateTestID(),
		}

		payload, err := warpQuery.Marshal()
		if err != nil {
			return err
		}

		uwm, err := warp.NewUnsignedMessage(uint32(1), ids.Empty, payload)
		if err != nil {
			return err
		}
		wm, err := warp.NewMessage(uwm, &warp.BitSetSignature{})
		if err != nil {
			return err
		}

		_, _, err = sendAndWait(ctx, wm, &actions.Query{}, cli, bcli, factory, true)

		return err
	},
}
