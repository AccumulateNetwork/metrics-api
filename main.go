package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AccumulateNetwork/metrics-api/accumulate"
	"github.com/AccumulateNetwork/metrics-api/api"
	"github.com/AccumulateNetwork/metrics-api/global"
	"github.com/AccumulateNetwork/metrics-api/schema"
	"github.com/labstack/gommon/log"
)

const ACCUMULATE_API = "https://mainnet.accumulatenetwork.io/v2"
const ACCUMULATE_CLIENT_TIMEOUT = 5
const API_PORT = 8082
const STAKING_DATA_ACCOUNT = "acc://staking.acme/registered"

func main() {

	global.StakingRecords = &schema.StakingRecords{}

	client := accumulate.NewAccumulateClient(ACCUMULATE_API, ACCUMULATE_CLIENT_TIMEOUT)

	die := make(chan bool)
	go getACMEStats(client, die)

	fmt.Println("Starting Accumulate Staking API at port", API_PORT)
	log.Fatal(api.StartAPI(API_PORT))
}

func getACMEStats(client *accumulate.AccumulateClient, die chan bool) {

	for {

		select {
		default:

			stakingData, err := client.QueryDataSet(&accumulate.Params{URL: STAKING_DATA_ACCOUNT, Count: 10000, Start: 0, Expand: true})
			if err != nil {
				log.Error(err)
			}

			log.Info("received ", len(stakingData.Items), " data entries from ", STAKING_DATA_ACCOUNT)

			for _, entry := range stakingData.Items {

				entryData, err := hex.DecodeString(entry.Entry.Data[0])
				if err != nil {
					log.Error(err)
					continue
				}

				stRecord, err := accumulate.ParseStakingRecord(entryData)
				if err != nil {
					log.Error(err)
					continue
				}

				// check if record with this identity already exists
				exists := global.SearchStakingRecordByIdentity(stRecord.Identity)

				// if not found, append new record
				if exists == nil {
					log.Info("added staking record for: ", stRecord.Identity)
					global.StakingRecords.Items = append(global.StakingRecords.Items, stRecord)
					continue
				}

				log.Info("updated staking record for: ", stRecord.Identity)
				*exists = *stRecord

			}

			log.Info("total staking records: ", len(global.StakingRecords.Items))

			time.Sleep(time.Duration(15) * time.Minute)

		case <-die:
			return
		}

	}

}
