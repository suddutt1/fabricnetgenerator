package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//GenerateBuildAndJoinChannelScriptV2x generates build and join channel script for fabric 2.2.x
func GenerateBuildAndJoinChannelScriptV2x(configBytes []byte) bool {
	fmt.Println("Generating fabric 2.2.x complaint build and join channel scripts ")
	var netConf NetworkConfig
	err := json.Unmarshal(configBytes, &netConf)
	if err != nil {
		fmt.Println("Error is generating build and join script")
		return false
	}
	orgMap := netConf.GetOrgMap()
	orderer := fmt.Sprintf("%s:7050", netConf.Orderer.GetFQDNList()[0])
	var outputBuf bytes.Buffer
	for _, channelInfo := range netConf.Channels {

		header := fmt.Sprintf(buildAndJoinChannelScriptHeaderP1, channelInfo.Name, channelInfo.Name)
		outputBuf.WriteString(header)
		//Channel register
		channelReg := fmt.Sprintf(channelRegister, channelInfo.Orgs[0], orderer, channelInfo.Name)
		outputBuf.WriteString(channelReg)
		//Now run joinchannels

		for _, org := range channelInfo.Orgs {
			outputBuf.WriteString(fmt.Sprintf("\n# Joining %s for org peers of %s\n", channelInfo.Name, org))
			orgDetails, _ := orgMap[org]
			for index := 0; index < orgDetails.Peers; index++ {
				joinChan := fmt.Sprintf(joinChannel, orgDetails.Name, index)
				outputBuf.WriteString(joinChan)
			}
			//Perform the anchor peer update
			anchorPeerUpd := fmt.Sprintf(anchoPeerUpdate, org, org, orderer, channelInfo.Name, orgDetails.MSPID)
			outputBuf.WriteString(anchorPeerUpd)
		}

	}
	ioutil.WriteFile("./buildandjoinchannel.sh", outputBuf.Bytes(), 0777)
	return true
}

const buildAndJoinChannelScriptHeaderP1 = `#!/bin/bash

echo "Building channel for %s" 
export CHANNEL_NAME="%s"

`
const channelRegister = `

#Register the channel with orderer

. setpeer.sh %s peer0
peer channel create -o %s -c $CHANNEL_NAME -f ./%s.tx --tls true --cafile $ORDERER_CA -t 1000s
`

const joinChannel = `

. setpeer.sh %s peer%d
peer channel join -b $CHANNEL_NAME.block
`
const anchoPeerUpdate = `

#Update the anchor peers for org %s
. setpeer.sh %s peer0
peer channel update -o  %s -c $CHANNEL_NAME -f ./%s%sAnchor.tx --tls --cafile $ORDERER_CA 
`
