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

			copier.Copy(&acme, acmeData.Data)

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

			// parse staking data entries
			for _, entry := range stakingData.Items {

				entryData, err := hex.DecodeString(entry.Entry.Data[0])
				if err != nil {
					log.Error(err)
					continue
				}

				stRecord, err := schema.ParseStakingRecord(entryData)
				if err != nil {
					log.Error(err)
					continue
				}

				// fill entry hash
				stRecord.EntryHash = entry.EntryHash

				// check if record with this identity already exists
				exists := store.SearchStakingRecordByIdentity(stRecord.Identity)

				// if not found, append new record
				if exists == nil {
					log.Debug("added staking record for: ", stRecord.Identity)
					store.StakingRecords.Items = append(store.StakingRecords.Items, stRecord)
					continue
				}

				log.Debug("updated staking record for: ", stRecord.Identity)
				*exists = *stRecord

			}

			log.Info("total staking records: ", len(store.StakingRecords.Items))

			snapshot := &schema.StakingRecords{}
			copier.Copy(&snapshot.Items, store.StakingRecords.Items)

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

			copier.Copy(&store.StakingRecords.Items, snapshot.Items)

			now := time.Now()
			store.UpdatedAt = &now

			time.Sleep(time.Duration(10) * time.Minute)

		case <-die:
			return
		}

	}

}
