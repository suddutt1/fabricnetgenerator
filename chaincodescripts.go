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
		shFileInstall, _ := os.Create(path + ccID + "_install.sh")
		shFileInstall.WriteString("#!/bin/bash\n")
		shFileUpdateCC, _ := os.Create(path + ccID + "_update.sh")
		shFileUpdateCC.WriteString("#!/bin/bash\n")
		shFileUpdateCC.WriteString("if [[ ! -z \"$1\" ]]; then  \n")
		policy := ""
		for _, participant := range participants {
			peerCount := peerCountMap[getString(participant)]
			for index := 0; index < peerCount; index++ {
				lineToWrite := fmt.Sprintf(". setpeer.sh %s peer%d \n", participant, index)
				shFileInstall.WriteString(lineToWrite)
				shFileUpdateCC.WriteString("\t" + lineToWrite)
				exeCommand := fmt.Sprintf("peer chaincode install -n %s -v %s -p %s\n", ccID, version, src)
				shFileInstall.WriteString(exeCommand)
				exeUpdCommand := fmt.Sprintf("peer chaincode install -n %s -v %s -p %s\n", ccID, "$1", src)
				shFileUpdateCC.WriteString("\t" + exeUpdCommand)
			}
			policy = policy + ",'" + (mspMap[getString(participant)]) + ".member'"
		}
		runes := []rune(policy)
		finalPolicy := string(runes[1:])
		instCommand := fmt.Sprintf("peer chaincode instantiate -o %s:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C %s -n %s -v %s -c '{\"Args\":[\"init\",\"\"]}' -P \" OR( %s ) \" \n", ordererFDQN, channelName, ccID, version, finalPolicy)
		shFileInstall.WriteString(instCommand)
		updateCommand := fmt.Sprintf("\tpeer chaincode upgrade -o %s:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C %s -n %s -v %s -c '{\"Args\":[\"init\",\"\"]}' -P \" OR( %s ) \" \n", ordererFDQN, channelName, ccID, "$1", finalPolicy)
		shFileUpdateCC.WriteString(updateCommand)
		shFileUpdateCC.WriteString("else\n")
		shFileUpdateCC.WriteString("\techo \". " + ccID + "_updchain.sh  <Version Number>\" \n")
		shFileUpdateCC.WriteString("fi\n")
		shFileInstall.Close()
		shFileUpdateCC.Close()
	}

	//instCommand =
	//     peer chaincode instantiate -o orderer.kg.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $1 -v $2 -c '{"Args":["init",""]}' -P "OR ('RawMaterialDepartmentMSP.member','ManufacturingDepartmentMSP.member','DistributionCenterMSP.member','DistributionCenterMSP.member')"

	return true
}
