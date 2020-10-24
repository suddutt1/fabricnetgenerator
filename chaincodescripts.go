package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//GenerateChainCodeScripts generate chaincode scipts
func GenerateChainCodeScripts(config []byte, path string) bool {
	fmt.Println("Generating config scripts")
	fileNames := make([]string, 0)
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	version, _ := dataMapContainer["fabricVersion"].(string)
	if strings.HasPrefix(version, "2.2") {
		return GenerateNewLifecycleCCScripts(config)

	}
	//Build the msp info
	mspMap := make(map[string]string)
	peerCountMap := make(map[string]int)
	ordererConfig := getMap(dataMapContainer["orderers"])
	ordererFDQN := getString(ordererConfig["ordererHostname"]) + "." + getString(ordererConfig["domain"])
	if ifExists(ordererConfig, "type") && ifExists(ordererConfig, "haCount") && (getString(ordererConfig["type"]) == "kafka" || getString(ordererConfig["type"]) == "raft") {
		ordererFDQN = getString(ordererConfig["ordererHostname"]) + "0." + getString(ordererConfig["domain"])
	}
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
		channelName := fmt.Sprintf("%s", strings.ToLower((getString(chainCodeConfig["channelName"]))))
		participants, particpantExists := chainCodeConfig["participants"].([]interface{})
		if !particpantExists {
			fmt.Printf("No participants \n")
			return false
		}
		instShFileName := path + ccID + "_install.sh"
		fileNames = append(fileNames, instShFileName)
		shFileInstall, _ := os.Create(instShFileName)
		shFileInstall.WriteString("#!/bin/bash\n")
		updShFileName := path + ccID + "_update.sh"
		fileNames = append(fileNames, updShFileName)
		shFileUpdateCC, _ := os.Create(updShFileName)
		shFileUpdateCC.WriteString("#!/bin/bash\n")
		shFileUpdateCC.WriteString("if [[ ! -z \"$1\" ]]; then  \n")
		policy := ""
		for _, participant := range participants {
			peerCount := peerCountMap[getString(participant)]
			for index := 0; index < peerCount; index++ {
				lineToWrite := fmt.Sprintf(". setpeer.sh %s peer%d \n", participant, index)
				setChannel := fmt.Sprintf("export CHANNEL_NAME=\"%s\"\n", channelName)
				shFileInstall.WriteString(lineToWrite)
				shFileInstall.WriteString(setChannel)
				shFileUpdateCC.WriteString("\t" + lineToWrite)
				shFileUpdateCC.WriteString(setChannel)
				exeCommand := fmt.Sprintf("peer chaincode install -n %s -v %s -l golang -p  %s\n", ccID, version, src)
				shFileInstall.WriteString(exeCommand)
				exeUpdCommand := fmt.Sprintf("peer chaincode install -n %s -v %s -l golang -p  %s\n", ccID, "$1", src)
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
	for _, fileName := range fileNames {
		os.Chmod(fileName, 0777)
	}
	return true
}
