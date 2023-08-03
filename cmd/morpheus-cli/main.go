// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// "morpheus-cli" implements morpheusvm client operation interface.
package main

import (
	"os"

	"github.com/ava-labs/hypersdk/utils"
	"github.com/bianyuanop/oraclevm/cmd/morpheus-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		utils.Outf("{{red}}morpheus-cli exited with error:{{/}} %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
