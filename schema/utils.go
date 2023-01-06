package schema

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

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
