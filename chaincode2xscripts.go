package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//GenerateNewLifecycleCCScripts generated version 2.2.x compliant chaincode scripts
func GenerateNewLifecycleCCScripts(confBytes []byte) bool {
	fmt.Println("Generating 2.2.x compatible cc scripts")
	var netConf NetworkConfig
	err := json.Unmarshal(confBytes, &netConf)
	if err != nil {
		fmt.Println("Parsing error in network configuration")
		return false
	}
	orgMap := netConf.GetOrgMap()
	orderer := fmt.Sprintf("%s:7050", netConf.Orderer.GetFQDNList()[0])
	for _, ccDetails := range netConf.Chaincodes {
		var sfBuf bytes.Buffer
		var updBuf bytes.Buffer
		part1 := fmt.Sprintf(ccInstallPart1, ccDetails.Channel, ccDetails.CCID, 1)

		sfBuf.WriteString(part1)
		updPart1 := fmt.Sprintf(ccInstallPart2, ccDetails.Channel, ccDetails.CCID, ccDetails.Participants[0])
		updBuf.WriteString(updPart1)
		instShFileName := fmt.Sprintf("./%s_install.sh", ccDetails.CCID)
		updateShFileName := fmt.Sprintf("./%s_update.sh", ccDetails.CCID)

		shFileInstall, _ := os.Create(instShFileName)
		shFileUpdate, _ := os.Create(updateShFileName)

		//Package the chaincode
		packageStr := fmt.Sprintf(ccInstallPartCCPakage, ccDetails.Participants[0], ccDetails.SrcPath)
		sfBuf.WriteString(packageStr)
		updBuf.WriteString(packageStr)

		var peerConnBuf bytes.Buffer
		for _, participantOrg := range ccDetails.Participants {
			orgDetails := orgMap[participantOrg]
			for peerIndex := 0; peerIndex < orgDetails.Peers; peerIndex++ {
				installPkg := fmt.Sprintf(ccInstallCC, participantOrg, peerIndex)
				sfBuf.WriteString(installPkg)
				updBuf.WriteString(installPkg)

			}
			peerConnBuf.WriteString(fmt.Sprintf(" --peerAddresses peer0.%s:7051 --tlsRootCertFiles ${%s_PEER0_CA} ", orgDetails.Domain, strings.ToUpper(orgDetails.Name)))

		}
		endorseentPolicy := ccDetails.GetCCEndorsementPolicy(orgMap)

		for _, participantOrg := range ccDetails.Participants {
			approve := fmt.Sprintf(ccApprove, participantOrg, participantOrg, participantOrg, participantOrg, orderer, endorseentPolicy)
			sfBuf.WriteString(approve)
			updBuf.WriteString(approve)

		}
		//Commit phase
		ccComit := fmt.Sprintf(ccCommit, peerConnBuf.String(), orderer, endorseentPolicy)
		sfBuf.WriteString(ccComit)
		updBuf.WriteString(ccComit)
		//Query the commited chaincode
		for _, participantOrg := range ccDetails.Participants {
			querycommitted := fmt.Sprintf(ccQueryCommit, participantOrg, participantOrg)
			sfBuf.WriteString(querycommitted)
			updBuf.WriteString(querycommitted)
		}
		//Now init the chaincode
		ccInitInvoke := fmt.Sprintf(ccInvokeInit, orderer)
		sfBuf.WriteString(ccInitInvoke)
		updBuf.WriteString(ccInitInvoke)

		shFileInstall.WriteString(sfBuf.String())
		shFileUpdate.WriteString(updBuf.String())
		shFileInstall.Close()
		shFileUpdate.Close()
		os.Chmod(instShFileName, 0777)
		os.Chmod(updateShFileName, 0777)
	}
	return true

}

const ccInstallPart1 = `
#!/bin/sh
export CHANNEL_NAME="%s"
export CC_NAME="%s"
export CC_VERSION="%d.0"
export CC_SEQ=1

`
const ccInstallPart2 = `
#!/bin/sh
export CHANNEL_NAME="%s"
export CC_NAME="%s"
export CC_VERSION="2.0"
export CC_SEQ=2

# Retrive last version 
. setpeer.sh  %s peer0

peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name ${CC_NAME} >&clog.txt
export SEQ=$(tail -n 1 clog.txt | awk -F ',' '{ print $2 }'| awk -F ':' '{print $2}')
export NEWSEQ=$((SEQ+1))
export CC_VERSION="${NEWSEQ}.0"
export CC_SEQ=$NEWSEQ
echo "Going to upgrade ${CC_NAME} on ${CHANNEL_NAME} to version ${CC_VERSION} sequence ${CC_SEQ} "

`

const ccInstallPartCCPakage = `
. setpeer.sh  %s peer0

# Package the chaincode 
cd chaincode/%s
peer lifecycle chaincode package /opt/ws/${CC_NAME}.tar.gz  --path .  --lang golang --label ${CC_NAME}_${CC_VERSION}
cd /opt/ws

`

const ccInstallCC = `

# Install the chaincode package 
. setpeer.sh %s peer%d
peer lifecycle chaincode install ${CC_NAME}.tar.gz

`
const ccApprove = `

# Approve Organization %s 

. setpeer.sh %s peer0
peer lifecycle chaincode queryinstalled >&cq%slog.txt
PACKAGE_ID=$(sed -n "/${CC_NAME}_${CC_VERSION}/{s/^Package ID: //; s/, Label:.*$//; p;}" cq%slog.txt)
echo $PACKAGE_ID
peer lifecycle chaincode approveformyorg -o %s  --tls --cafile $ORDERER_CA --channelID $CHANNEL_NAME --name $CC_NAME --version $CC_VERSION --package-id $PACKAGE_ID --sequence $CC_SEQ --init-required  --signature-policy " %s " 

`

const ccCommit = `

#Commit chaincode installation 

export PEER_CONN="%s"
peer lifecycle chaincode commit -o %s  --tls --cafile $ORDERER_CA --channelID $CHANNEL_NAME --name $CC_NAME --version $CC_VERSION --sequence $CC_SEQ --init-required --signature-policy " %s" $PEER_CONN

`
const ccQueryCommit = `

#Query commited in Org %s
. setpeer.sh %s peer0
peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name ${CC_NAME}

`
const ccInvokeInit = `
sleep 2
#Invoke init
peer chaincode invoke -o %s  --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n ${CC_NAME} $PEER_CONN --isInit -c '{"Args":[""]}'


`
