// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/crypto"
	"github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"
	"github.com/bianyuanop/oraclevm/actions"
	"github.com/bianyuanop/oraclevm/auth"
	brpc "github.com/bianyuanop/oraclevm/rpc"
	"github.com/spf13/cobra"
)

var spamCmd = &cobra.Command{
	Use: "spam",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var runSpamCmd = &cobra.Command{
	Use: "run",
	RunE: func(*cobra.Command, []string) error {
		var bclient *brpc.JSONRPCClient
		return handler.Root().Spam(maxTxBacklog, randomRecipient,
			func(uri string, networkID uint32, chainID ids.ID) {
				bclient = brpc.NewJSONRPCClient(uri, networkID, chainID)
			},
			func(pk crypto.PrivateKey) chain.AuthFactory {
				return auth.NewED25519Factory(pk)
			},
			func(choice int, address string) (uint64, error) {
				balance, err := bclient.Balance(context.TODO(), address)
				if err != nil {
					return 0, err
				}
				hutils.Outf(
					"%d) {{cyan}}address:{{/}} %s {{cyan}}balance:{{/}} %s %s\n",
					choice,
					address,
					handler.Root().ValueString(ids.Empty, balance),
					handler.Root().AssetString(ids.Empty),
				)
				return balance, err
			},
			func(ctx context.Context, chainID ids.ID) (chain.Parser, error) {
				return bclient.Parser(ctx)
			},
			func(pk crypto.PublicKey, amount uint64) chain.Action {
				return &actions.Transfer{
					To:    pk,
					Value: amount,
				}
			},
			func(cli *rpc.JSONRPCClient, pk crypto.PrivateKey) func(context.Context, uint64) error {
				return func(ictx context.Context, count uint64) error {
					_, _, err := sendAndWait(ictx, nil, &actions.Transfer{
						To:    pk.PublicKey(),
						Value: count, // prevent duplicate txs
					}, cli, bclient, auth.NewED25519Factory(pk), false)
					return err
				}
			},
		)
	},
}
