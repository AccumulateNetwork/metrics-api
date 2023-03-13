package store

import (
	"strings"

	"github.com/AccumulateNetwork/metrics-api/schema"
)

// SearchStakingRecordByIdentity searches staking record by Identity (case insensitive)
func SearchStakingRecordByIdentity(identity string, records []*schema.StakingRecord) *schema.StakingRecord {

	for _, r := range records {
		if strings.EqualFold(r.Identity, identity) {
			return r
		}
	}

	return nil

}

// SearchValidatorByIdentity searches validator by Identity (case insensitive)
func SearchValidatorByIdentity(identity string, validators []*schema.Validator) *schema.Validator {

	for _, r := range validators {
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

// GetValidators returns validators
func GetValidators() *schema.Validators {

	res := &schema.Validators{}

	// get all validators
	for _, r := range StakingRecords.Items {
		if r.Type == "coreValidator" || r.Type == "coreFollower" || r.Type == "stakingValidator" {
			validator := &schema.Validator{
				Type:               r.Type,
				Identity:           r.Identity,
				Stake:              r.Stake,
				Rewards:            r.Rewards,
				Balance:            r.Balance,
				AcceptingDelegates: r.AcceptingDelegates,
				TotalStaked:        r.Balance,
			}
			res.Items = append(res.Items, validator)
		}
	}

	// add delegated balances
	for _, r := range StakingRecords.Items {
		if r.Type == "delegated" {
			exists := SearchValidatorByIdentity(r.Delegate, res.Items)
			if exists.AcceptingDelegates == "yes" {
				exists.TotalStaked += r.Balance
			}
		}
	}

	return res

}
