package schema

import (
	"encoding/json"
	"sort"
	"strings"

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

// ParseStakingRecord parses Accumulate staking entry data V2 into struct and validates it
func ParseStakingRecordV2(entry []byte) (*StakingRecordV2, error) {

	var err error

	// unmarshal data entry into staking record V2
	res := &StakingRecordV2{}
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

// ParseTokenRecord parses Accumulate token entry data into struct and validates it
func ParseTokenRecord(entry []byte) (*Token, error) {

	var err error

	// unmarshal data entry into token
	res := &Token{}
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
				return strings.ToLower(sr.Items[i].Identity) > strings.ToLower(sr.Items[j].Identity)
			} else {
				return strings.ToLower(sr.Items[i].Identity) < strings.ToLower(sr.Items[j].Identity)
			}
		})
	}

}

// Sort sorts validators by `sort` field and `order` ordering
func (v *Validators) Sort(sorting string, order string) {

	switch sorting {
	case "balance":
		sort.Slice(v.Items[:], func(i, j int) bool {
			if order == AlternativeOrder {
				return v.Items[i].Balance > v.Items[j].Balance
			} else {
				return v.Items[i].Balance < v.Items[j].Balance
			}
		})
	case "identity":
		sort.Slice(v.Items[:], func(i, j int) bool {
			if order == AlternativeOrder {
				return strings.ToLower(v.Items[i].Identity) > strings.ToLower(v.Items[j].Identity)
			} else {
				return strings.ToLower(v.Items[i].Identity) < strings.ToLower(v.Items[j].Identity)
			}
		})
	case "totalStaked":
		sort.Slice(v.Items[:], func(i, j int) bool {
			if order == AlternativeOrder {
				return v.Items[i].TotalStaked > v.Items[j].TotalStaked
			} else {
				return v.Items[i].TotalStaked < v.Items[j].TotalStaked
			}
		})
	}

}
