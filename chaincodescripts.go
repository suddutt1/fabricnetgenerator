package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func GenerateChainCodeScripts(config []byte, path string) bool {
	fmt.Println("Generating config scripts")
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	//Build the msp info
	mspMap := make(map[string]string)
	peerCountMap := make(map[string]int)
	ordererConfig := getMap(dataMapContainer["orderers"])
	ordererFDQN := getString(ordererConfig["ordererHostname"]) + "." + getString(ordererConfig["domain"])
	orgs, orgsExists := dataMapContainer["orgs"].([]interface{})
	if !orgsExists {
		fmt.Println("No organizations specified")
		return false
	}
	for _, org := range orgs {
		orgConfig := getMap(org)
		name := getString(orgConfig["name"])
		mspId := getString(orgConfig["mspID"])
		mspMap[name] = mspId
		peerCountMap[name] = getNumber(orgConfig["peerCount"])
	}
	chainCodes, chainExists := dataMapContainer["chaincodes"].([]interface{})
	if !chainExists {
		fmt.Println("No chain codes defined")
		return false
	}
	shFile, _ := os.Create(path + "installcc.sh")
	shFile.WriteString("#!/bin/bash\n")
	for _, ccInfo := range chainCodes {
		chainCodeConfig := getMap(ccInfo)
		ccID := getString(chainCodeConfig["ccid"])
		version := getString(chainCodeConfig["version"])
		src := getString(chainCodeConfig["src"])
		channelName := fmt.Sprintf("%schannel", strings.ToLower((getString(chainCodeConfig["channelName"]))))
		participants, particpantExists := chainCodeConfig["participants"].([]interface{})
		if !particpantExists {
			fmt.Printf("No participants \n")
			return false
		}
		policy := ""
		for _, participant := range participants {
			peerCount := peerCountMap[getString(participant)]
			for index := 0; index < peerCount; index++ {
				lineToWrite := fmt.Sprintf(". setpeer.sh %s peer%d \n", participant, index)
				shFile.WriteString(lineToWrite)
				exeCommand := fmt.Sprintf("peer chaincode install -n %s -v %s -p %s\n", ccID, version, src)
				shFile.WriteString(exeCommand)
			}
			policy = policy + ",'" + (mspMap[getString(participant)]) + ".member'"
		}
		runes := []rune(policy)
		finalPolicy := string(runes[1:])
		instCommand := fmt.Sprintf("peer chaincode instantiate -o %s:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C %s -n %s -v %s -c '{\"Args\":[\"init\",\"\"]}' -P \" OR( %s ) \" \n", ordererFDQN, channelName, ccID, version, finalPolicy)
		shFile.WriteString(instCommand)
	}

	//instCommand =
	//     peer chaincode instantiate -o orderer.kg.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $1 -v $2 -c '{"Args":["init",""]}' -P "OR ('RawMaterialDepartmentMSP.member','ManufacturingDepartmentMSP.member','DistributionCenterMSP.member','DistributionCenterMSP.member')"
	shFile.Close()
	return true
}
