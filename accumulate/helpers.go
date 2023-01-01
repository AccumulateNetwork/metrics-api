package accumulate

import (
	"encoding/json"
	"strings"

	"github.com/AccumulateNetwork/metrics-api/global"
	"github.com/AccumulateNetwork/metrics-api/schema"
	"github.com/go-playground/validator/v10"
)

func ParseStakingRecord(entry []byte) (*schema.StakingRecord, error) {

	var err error

	// unmarshal data entry into staking record
	res := &schema.StakingRecord{}
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

func SearchStakingRecordByIdentity(identity string) *schema.StakingRecord {

	for _, r := range global.StakingRecords.Items {
		if strings.EqualFold(r.Identity, identity) {
			return r
		}
	}

	return nil

}
