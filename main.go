package main

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/AccumulateNetwork/metrics-api/accumulate"
	"github.com/AccumulateNetwork/metrics-api/api"
	"github.com/AccumulateNetwork/metrics-api/schema"
	"github.com/AccumulateNetwork/metrics-api/store"
	"github.com/jinzhu/copier"
	"github.com/labstack/gommon/log"
)

const ACCUMULATE_API = "https://mainnet.accumulatenetwork.io/v2"
const ACCUMULATE_CLIENT_TIMEOUT = 5
const API_PORT = 8082
const ACME_TOKEN_ISSUER = "acc://acme"
const STAKING_DATA_ACCOUNT = "acc://staking.acme/registered"
const STAKING_PAGESIZE = 10000

func main() {

	store.StakingRecords = &schema.StakingRecords{}

	client := accumulate.NewAccumulateClient(ACCUMULATE_API, ACCUMULATE_CLIENT_TIMEOUT)

	die := make(chan bool)
	go getStats(client, die)

	log.Fatal(api.StartAPI(API_PORT))
}

func getStats(client *accumulate.AccumulateClient, die chan bool) {

	for {

		select {
		default:

			acme := &schema.ACME{}

			acmeData, err := client.QueryToken(&accumulate.Params{URL: ACME_TOKEN_ISSUER})
			if err != nil {
				log.Error(err)
			}

			copier.Copy(&acme, &acmeData.Data)

			acme.Total, err = strconv.ParseInt(acmeData.Data.Issued, 10, 64)
			if err != nil {
				log.Error(err)
			}

			acme.Max, err = strconv.ParseInt(acmeData.Data.SupplyLimit, 10, 64)
			if err != nil {
				log.Error(err)
			}

			store.ACME = acme

			stakingData, err := client.QueryDataSet(&accumulate.Params{URL: STAKING_DATA_ACCOUNT, Count: STAKING_PAGESIZE, Start: 0, Expand: true})
			if err != nil {
				log.Error(err)
			}

			log.Info("received ", len(stakingData.Items), " data entries from ", STAKING_DATA_ACCOUNT)

			snapshot := &schema.StakingRecords{}
			copier.Copy(&snapshot.Items, &store.StakingRecords.Items)

			// parse staking data entries
			for _, entry := range stakingData.Items {

				entryData, err := hex.DecodeString(entry.Entry.Data[0])
				if err != nil {
					log.Error(err)
					continue
				}

				stRecordV2 := &schema.StakingRecordV2{}

				stRecord, err := schema.ParseStakingRecord(entryData)
				if err != nil {
					// fallback
					stRecordV2, err = schema.ParseStakingRecordV2(entryData)
					if err != nil {
						log.Error(err, " ", entry.EntryHash)
						continue
					}
				}

				if stRecordV2.Identity != "" {
					// v2 case
					// check if record with this identity and stake already exists
					exists := store.SearchStakingRecordByIdentity(stRecordV2.Identity, snapshot.Items)

					if exists != nil {
						store.RemoveStakingRecordsByIdentity(stRecordV2.Identity, snapshot.Items)
					}

					for _, account := range stRecordV2.Accounts {

						// declare stRecord for each account
						stRecord = &schema.StakingRecord{}

						// fill identity
						stRecord.Identity = stRecordV2.Identity

						// fill entry hash
						stRecord.EntryHash = entry.EntryHash

						// fill fields
						stRecord.AcceptingDelegates = account.AcceptingDelegates
						stRecord.Delegate = account.Delegate
						stRecord.Rewards = account.Rewards
						stRecord.Stake = account.Stake
						stRecord.Type = account.Type

						log.Debug("added staking record for: ", stRecord.Identity)
						snapshot.Items = append(snapshot.Items, stRecord)
					}
				} else {
					// v1 case

					// fill entry hash
					stRecord.EntryHash = entry.EntryHash

					// check if record with this identity and stake already exists
					exists := store.SearchStakingRecordByIdentity(stRecord.Identity, snapshot.Items)

					// if not found, append new record
					if exists == nil {
						log.Debug("added staking record for: ", stRecord.Identity)
						snapshot.Items = append(snapshot.Items, stRecord)
						continue
					}

					log.Debug("updated staking record for: ", stRecord.Identity)
					*exists = *stRecord
				}

			}

			log.Info("total staking records: ", len(stakingData.Items))

			// get ACME balances of stakers
			for _, record := range snapshot.Items {

				balance, err := client.QueryTokenAccount(&accumulate.Params{URL: record.Stake})
				if err != nil {
					log.Error(err)
					log.Info(record)
					continue
				}

				record.Balance, err = strconv.ParseInt(balance.Data.Balance, 10, 64)
				if err != nil {
					log.Error(err)
					continue
				}

			}

			copier.Copy(&store.StakingRecords.Items, &snapshot.Items)

			foundationTotalBalance := int64(0)

			// get ACME balances of the foundation accounts
			for _, foundationAcc := range store.FoundationAccounts {

				balance, err := client.QueryTokenAccount(&accumulate.Params{URL: foundationAcc})
				if err != nil {
					log.Error(err)
					continue
				}

				parsedBalance, err := strconv.ParseInt(balance.Data.Balance, 10, 64)
				if err != nil {
					log.Error(err)
					continue
				}

				foundationTotalBalance += parsedBalance

			}

			log.Info("foundation total balance: ", foundationTotalBalance)

			copier.Copy(&store.FoundationTotalBalance, &foundationTotalBalance)

			now := time.Now()
			store.UpdatedAt = &now

			time.Sleep(time.Duration(1) * time.Minute)

		case <-die:
			return
		}

	}

}
