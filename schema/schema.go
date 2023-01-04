package schema

type StakingRecord struct {
	Type               string `json:"type" validate:"required"`
	Status             string `json:"status"`
	Identity           string `json:"identity" validate:"required"`
	Stake              string `json:"stake" validate:"required"`
	Rewards            string `json:"rewards" validate:"required"`
	Delegate           string `json:"delegate"`
	AcceptingDelegates string `json:"acceptingDelegates"`
	EntryHash          string `json:"entryHash"`
	Balance            int64  `json:"balance"`
}

type StakingRecords struct {
	Items []*StakingRecord `json:"items"`
}

type ACME struct {
	Issued      string `json:"issued"`
	SupplyLimit string `json:"supplyLimit"`
	Symbol      string `json:"symbol"`
	Precision   int64  `json:"precision"`
}

type ValidatorsNumber struct {
	CoreValidator    int64 `json:"coreValidator"`
	CoreFollower     int64 `json:"coreFollower"`
	StakingValidator int64 `json:"stakingValidator"`
	Delegated        int64 `json:"delegated"`
	Pure             int64 `json:"pure"`
}
