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

// GetTotalStake returns total staked ACME
func GetTotalStake() int64 {

	total := int64(0)

	for _, r := range StakingRecords.Items {
		total += r.Balance
	}

	return total

}

// GetValidatorsNumber returns number of validators
func GetValidatorsNumber() *schema.ValidatorsNumber {

	res := &schema.ValidatorsNumber{}

	for _, r := range StakingRecords.Items {
		switch r.Type {
		case "coreValidator":
			res.CoreValidator++
		case "coreFollower":
			res.CoreFollower++
		case "stakingValidator":
			res.StakingValidator++
		case "delegated":
			res.Delegated++
		default:
			res.Pure++
		}
	}

	return res

}
