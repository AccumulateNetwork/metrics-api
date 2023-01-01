package schema

type StakingRecord struct {
	Type               string `json:"type" validate:"required"`
	Status             string `json:"status"`
	Identity           string `json:"identity" validate:"required"`
	Stake              string `json:"stake" validate:"required"`
	Rewards            string `json:"rewards" validate:"required"`
	Delegate           string `json:"delegate"`
	AcceptingDelegates string `json:"acceptingDelegates"`
}

type StakingRecords struct {
	Items []*StakingRecord `json:"items"`
}
