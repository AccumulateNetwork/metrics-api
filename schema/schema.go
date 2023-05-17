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
type StakingRecordV2 struct {
	Status   string                    `json:"status"`
	Identity string                    `json:"identity" validate:"required"`
	Accounts []*StakingRecordV2Account `json:"accounts" validate:"required"`
}

type StakingRecordV2Account struct {
	Type               string `json:"type" validate:"required"`
	Stake              string `json:"url" validate:"required"`
	Rewards            string `json:"payout" validate:"required"`
	Delegate           string `json:"delegate"`
	AcceptingDelegates string `json:"acceptingDelegates"`
}

type StakingRecords struct {
	Items []*StakingRecord `json:"items"`
}

type Validator struct {
	Type               string `json:"type"`
	Identity           string `json:"identity"`
	Stake              string `json:"stake"`
	Rewards            string `json:"rewards"`
	Balance            int64  `json:"balance"`
	AcceptingDelegates string `json:"acceptingDelegates"`
	TotalStaked        int64  `json:"totalStaked"`
}

type Validators struct {
	Items []*Validator `json:"items"`
}

type ACME struct {
	Symbol    string `json:"symbol"`
	Precision int64  `json:"precision"`
	Total     int64  `json:"total"`
	Max       int64  `json:"max"`
}

type ValidatorsNumber struct {
	CoreValidator    int64 `json:"coreValidator"`
	CoreFollower     int64 `json:"coreFollower"`
	StakingValidator int64 `json:"stakingValidator"`
	Delegated        int64 `json:"delegated"`
	Pure             int64 `json:"pure"`
}

type Token struct {
	TokenIssuer string `json:"tokenIssuer" validate:"required,startswith=acc://"`
	Symbol      string `json:"symbol" validate:"required"`
	Logo        string `json:"logo" validate:"required,url"`
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required,url"`
}

type Tokens struct {
	Items []*Token `json:"items"`
}
