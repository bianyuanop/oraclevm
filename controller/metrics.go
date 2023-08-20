// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/bianyuanop/oraclevm/consts"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	transfer prometheus.Counter
	upload   prometheus.Counter
	query    prometheus.Counter
}

func newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
	m := &metrics{
		transfer: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "transfer",
			Help:      "number of transfer actions",
		}),
		upload: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "upload",
			Help:      "number of upload actions",
		}),
		query: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "query",
			Help:      "number of query actions",
		}),
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.transfer),

		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
