package global

import (
	"strings"

	"github.com/AccumulateNetwork/metrics-api/schema"
)

func SearchStakingRecordByIdentity(identity string) *schema.StakingRecord {

	for _, r := range StakingRecords.Items {
		if strings.EqualFold(r.Identity, identity) {
			return r
		}
	}

	return nil

}
