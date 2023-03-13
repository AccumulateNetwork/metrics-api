package store

import (
	"time"

	"github.com/AccumulateNetwork/metrics-api/schema"
)

var StakingRecords *schema.StakingRecords
var ACME *schema.ACME
var UpdatedAt *time.Time

var FoundationAccounts = []string{
	"acc://accumulate.acme/dev-block",
	"acc://accumulate.acme/factom-block",
	"acc://accumulate.acme/business/grants",
	"acc://accumulate.acme/core-dev/grants",
	"acc://accumulate.acme/ecosystem/grants",
	"acc://accumulate.acme/governance/grants",
	"acc://accumulate.acme/grant-block",
	"acc://accumulate.acme/stake",
	"acc://defi-growth-fund.acme/liquid-staking-rewards-boost",
}

var FoundationTotalBalance int64
