package store

import (
	"strings"

	"github.com/AccumulateNetwork/metrics-api/schema"
)

const ACC_PROTOCOL = "acc://"

// SearchStakingRecordByIdentity searches staking record by Identity (case insensitive)
func SearchStakingRecordByIdentity(identity string, records []*schema.StakingRecord) *schema.StakingRecord {

	for _, r := range records {
		if strings.EqualFold(r.Identity, identity) {
			return r
		}
	}

	return nil

}

// RemoveStakingRecordsByIdentity searches staking records by Identity (case insensitive) and removes them
func RemoveStakingRecordsByIdentity(identity string, records []*schema.StakingRecord) []*schema.StakingRecord {

	for i := 0; i < len(records); i++ {
		if strings.EqualFold(records[i].Identity, identity) {
			records = append(records[:i], records[i+1:]...)
			i--
		}
	}

	return records

}

// SearchStakingRecordByAccount searches staking record by Identity (case insensitive)
func SearchStakingRecordByAccount(stake string, records []*schema.StakingRecord) *schema.StakingRecord {

	for _, r := range records {
		if strings.EqualFold(r.Stake, stake) || strings.EqualFold(strings.ReplaceAll(r.Stake, ACC_PROTOCOL, ""), stake) {
			return r
		}
	}

	return nil

}

// SearchStakingRecordByIdentity searches staking record by Identity (case insensitive)
func SearchStakingRecordByIdentityAndAccount(identity string, stake string, records []*schema.StakingRecord) *schema.StakingRecord {

	for _, r := range records {
		if strings.EqualFold(r.Identity, identity) && strings.EqualFold(r.Stake, stake) {
			return r
		}
	}

	return nil

}

// SearchTokenByTokenIssuer searches staking record by Identity (case insensitive)
func SearchTokenByTokenIssuer(tokenIssuer string, records []*schema.Token) *schema.Token {

	for _, r := range records {
		if strings.EqualFold(r.TokenIssuer, tokenIssuer) {
			return r
		}
	}

	return nil

}

// SearchTokenBySymbol searches staking record by Symbol (case insensitive)
func SearchTokenBySymbol(symbol string, records []*schema.Token) *schema.Token {

	for _, r := range records {
		if strings.EqualFold(r.Symbol, symbol) {
			return r
		}
	}

	return nil

}

// SearchValidatorByIdentity searches validator by Identity (case insensitive)
func SearchValidatorByIdentity(identity string, validators []*schema.Validator) *schema.Validator {

	for _, r := range validators {
		if strings.EqualFold(r.Identity, identity) || strings.EqualFold(strings.ReplaceAll(r.Identity, ACC_PROTOCOL, ""), identity) {
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
			if exists != nil {
				exists.TotalStaked += r.Balance
			}
		}
	}

	return res

}
