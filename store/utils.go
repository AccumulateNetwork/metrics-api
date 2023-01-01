package store

import (
	"strings"

	"github.com/AccumulateNetwork/metrics-api/schema"
)

// SearchStakingRecordByIdentity searches staking record by Identity (case insensitive)
func SearchStakingRecordByIdentity(identity string) *schema.StakingRecord {

	for _, r := range StakingRecords.Items {
		if strings.EqualFold(r.Identity, identity) {
			return r
		}
	}

	return nil

}
