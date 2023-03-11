package schema

import (
	"encoding/json"
	"sort"

	"github.com/go-playground/validator/v10"
)

const DefaultOrder = "asc"
const AlternativeOrder = "desc"

// ParseStakingRecord parses Accumulate staking entry data into struct and validates it
func ParseStakingRecord(entry []byte) (*StakingRecord, error) {

	var err error

	// unmarshal data entry into staking record
	res := &StakingRecord{}
	if err = json.Unmarshal(entry, &res); err != nil {
		return nil, err
	}

	// validate staking record
	validate := validator.New()
	if err = validate.Struct(res); err != nil {
		return nil, err
	}

	return res, nil

}

// Sort sorts staking records by `sort` field and `order` ordering
func (sr *StakingRecords) Sort(sorting string, order string) {

	switch sorting {
	case "balance":
		sort.Slice(sr.Items[:], func(i, j int) bool {
			if order == AlternativeOrder {
				return sr.Items[i].Balance > sr.Items[j].Balance
			} else {
				return sr.Items[i].Balance < sr.Items[j].Balance
			}
		})
	case "identity":
		sort.Slice(sr.Items[:], func(i, j int) bool {
			if order == AlternativeOrder {
				return sr.Items[i].Identity > sr.Items[j].Identity
			} else {
				return sr.Items[i].Identity < sr.Items[j].Identity
			}
		})
	}

}
