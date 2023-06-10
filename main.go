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
const TOKENS_DATA_ACCOUNT = "acc://tokens.acme/list"
const ACCUMULATE_API_PAGESIZE = 100

func main() {

	store.StakingRecords = &schema.StakingRecords{}
	store.Tokens = &schema.Tokens{}

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

			// paging loop
			stakingData := &accumulate.QueryDataSetResponse{}
			flagFinished := false
			start := int64(0)

			for flagFinished != true {

				// query data set
				resp, err := client.QueryDataSet(&accumulate.Params{URL: STAKING_DATA_ACCOUNT, Count: ACCUMULATE_API_PAGESIZE, Start: start, Expand: true})
				if err != nil {
					log.Error(err)
					continue
				}

				// append entries
				stakingData.Items = append(stakingData.Items, resp.Items...)

				// next page
				start += ACCUMULATE_API_PAGESIZE

				// if pages ended, finish loop
				if resp.Start+ACCUMULATE_API_PAGESIZE >= resp.Total {
					flagFinished = true
				}

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
						snapshot.Items = store.RemoveStakingRecordsByIdentity(stRecordV2.Identity, snapshot.Items)
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

						log.Info("added staking record for: ", stRecord.Identity)
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

			tokensData, err := client.QueryDataSet(&accumulate.Params{URL: TOKENS_DATA_ACCOUNT, Count: ACCUMULATE_API_PAGESIZE, Start: 0, Expand: true})
			if err != nil {
				log.Error(err)
			}

			log.Info("received ", len(tokensData.Items), " data entries from ", TOKENS_DATA_ACCOUNT)

			tokenSnapshot := &schema.Tokens{}
			copier.Copy(&tokenSnapshot.Items, &store.Tokens.Items)

			// parse tokens data entries
			for _, entry := range tokensData.Items {

				entryData, err := hex.DecodeString(entry.Entry.Data[1])
				if err != nil {
					log.Error(err)
					continue
				}

				tokenRecord, err := schema.ParseTokenRecord(entryData)
				if err != nil {
					log.Error(err, " ", entry.EntryHash)
					continue
				}

				// check if record with this identity and stake already exists
				exists := store.SearchTokenByTokenIssuer(tokenRecord.TokenIssuer, tokenSnapshot.Items)

				// if not found, append new record
				if exists == nil {
					log.Debug("added token: ", tokenRecord.TokenIssuer)
					tokenSnapshot.Items = append(tokenSnapshot.Items, tokenRecord)
					continue
				}

				log.Debug("updated token: ", tokenRecord.TokenIssuer)
				*exists = *tokenRecord

			}

			copier.Copy(&store.Tokens.Items, &tokenSnapshot.Items)

			now := time.Now()
			store.UpdatedAt = &now

			time.Sleep(time.Duration(5) * time.Minute)

		case <-die:
			return
		}

	}

}
