package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
)

const _SetPeerTemplate = `
#!/bin/bash
export ORDERER_CA=/opt/ws/crypto-config/ordererOrganizations/{{.orderers.domain}}/msp/tlscacerts/tlsca.{{.orderers.domain}}-cert.pem
{{$primechannel := (index .channels 0).channelName }}
if [ $# -lt 2 ];then
	echo "Usage : . setpeer.sh {{range .orgs}}{{.name}}|{{end}} <peerid>"
fi
export peerId=$2
{{range .orgs}}
if [[ $1 = "{{.name}}" ]];then
	echo "Setting to organization {{.name}} peer "$peerId
	export CORE_PEER_ADDRESS=$peerId.{{.domain}}:7051
	export CORE_PEER_LOCALMSPID={{.mspID}}
	export CORE_PEER_TLS_CERT_FILE=/opt/ws/crypto-config/peerOrganizations/{{.domain}}/peers/$peerId.{{.domain}}/tls/server.crt
	export CORE_PEER_TLS_KEY_FILE=/opt/ws/crypto-config/peerOrganizations/{{.domain}}/peers/$peerId.{{.domain}}/tls/server.key
	export CORE_PEER_TLS_ROOTCERT_FILE=/opt/ws/crypto-config/peerOrganizations/{{.domain}}/peers/$peerId.{{.domain}}/tls/ca.crt
	export CORE_PEER_MSPCONFIGPATH=/opt/ws/crypto-config/peerOrganizations/{{.domain}}/users/Admin@{{.domain}}/msp
fi
{{end}}
	`
const _GenerateArtifactsTemplate = `
#!/bin/bash -e
export PWD={{ "pwd" | ToCMDString}}

export FABRIC_CFG_PATH=$PWD
export ARCH=$(uname -s)
export CRYPTOGEN=$PWD/bin/cryptogen
export CONFIGTXGEN=$PWD/bin/configtxgen

function generateArtifacts() {
	
	echo " *********** Generating artifacts ************ "
	echo " *********** Deleting old certificates ******* "
	
        rm -rf ./crypto-config
	
        echo " ************ Generating certificates ********* "
	
        $CRYPTOGEN generate --config=$FABRIC_CFG_PATH/crypto-config.yaml
        
        echo " ************ Generating tx files ************ "
	
		$CONFIGTXGEN -profile OrdererGenesis -outputBlock ./genesis.block
		{{range .channels}}{{$chName := .channelName }}{{$channelId:= $chName | ToLower }}
        $CONFIGTXGEN -profile {{print $chName "Channel"}} -outputCreateChannelTx ./{{print $channelId "channel.tx" }} -channelID {{ print $channelId "channel"}}
		{{end}}

}

generateArtifacts 

cd $PWD

`

const _GenerateArtifactsTemplateWithCA = `
#!/bin/bash -e
export PWD={{ "pwd" | ToCMDString}}

export FABRIC_CFG_PATH=$PWD
export ARCH=$(uname -s)
export CRYPTOGEN=$PWD/bin/cryptogen
export CONFIGTXGEN=$PWD/bin/configtxgen

function generateArtifacts() {
	
	echo " *********** Generating artifacts ************ "
	echo " *********** Deleting old certificates ******* "
	
        rm -rf ./crypto-config
	
        echo " ************ Generating certificates ********* "
	
        $CRYPTOGEN generate --config=$FABRIC_CFG_PATH/crypto-config.yaml
        
        echo " ************ Generating tx files ************ "
	
		$CONFIGTXGEN -profile OrdererGenesis -outputBlock ./genesis.block
		{{range .channels}}{{$chName := .channelName }}{{$channelId:= $chName | ToLower }}
        $CONFIGTXGEN -profile {{print $chName "Channel"}} -outputCreateChannelTx ./{{print $channelId "channel.tx" }} -channelID {{ print $channelId "channel"}}
		{{end}}

}
function generateDockerComposeFile(){
	OPTS="-i"
	if [ "$ARCH" = "Darwin" ]; then
		OPTS="-it"
	fi
	cp  docker-compose-template.yaml  docker-compose.yaml
	{{ range .orgs}}
	{{$orgName :=.name | ToUpper }}
	cd  crypto-config/peerOrganizations/{{.domain}}/ca
	PRIV_KEY=$(ls *_sk)
	cd ../../../../
	sed $OPTS "s/{{$orgName}}_PRIVATE_KEY/${PRIV_KEY}/g"  docker-compose.yaml
	{{end}}
}
generateArtifacts 
cd $PWD
generateDockerComposeFile
cd $PWD

`

const _BuildChannelScript = `
#!/bin/bash -e
{{ $orderer:= .ordererURL}}
{{ $root := . }}
{{range .channels}}
{{ $channelId := print .channelName "channel" | ToLower }}
echo "Building channel for {{print $channelId}}" 
{{$firstOrg := (index .orgs 0) }}
. setpeer.sh {{$firstOrg}} peer0
export CHANNEL_NAME="{{print $channelId }}"
peer channel create -o {{ print $orderer }} -c $CHANNEL_NAME -f ./{{print $channelId ".tx"}} --tls true --cafile $ORDERER_CA -t 10000
{{ range $index,$orgName :=.orgs}}{{$orgConfig :=  index $root $orgName }}
{{ range $i,$peerId:=$orgConfig.peerNames }}
. setpeer.sh {{$orgName}} {{$peerId}}
export CHANNEL_NAME="{{print $channelId }}"
peer channel join -b $CHANNEL_NAME.block
{{end}}{{end}}{{end}}
`
const _CleanUp = `
#!/bin/bash
echo "Clearing the old artifacts"
rm *.yaml
rm -rf crypto-config
rm *.block
rm *.tx
rm generateartifacts.sh
rm setenv.sh
rm setpeer.sh
rm buildandjoinchannel.sh
rm *_install.sh
rm *_update.sh

`
const _SetEnv = `
#!/bin/bash
export IMAGE_TAG="x86_64-{{.fabricVersion}}"
export TP_IMAGE_TAG="x86_64-{{.thirdPartyVersion}}"

`
const _DOTENV = `
COMPOSE_PROJECT_NAME=bc

`
const _DOWNLOAD_SCRIPTS = `
#!/bin/bash

export VERSION={{.fabricVersion}}
export ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
#Set MARCH variable i.e ppc64le,s390x,x86_64,i386
MARCH="x86_64"


: ${CA_TAG:="$MARCH-$VERSION"}
: ${FABRIC_TAG:="$MARCH-$VERSION"}

echo "===> Downloading platform binaries"
curl https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric/hyperledger-fabric/${ARCH}-${VERSION}/hyperledger-fabric-${ARCH}-${VERSION}.tar.gz | tar xz

`
const _VERSION_COMP_MAP = `
{
	"1.0.0":{ "fabricCore":"1.0.0","thirdParty":"1.0.0"},
	"1.0.4":{ "fabricCore":"1.0.4","thirdParty":"1.0.4"},
	"1.1.0":{ "fabricCore":"1.1.0","thirdParty":"1.0.6"}
	
}	

`

func ToCMDString(input string) string {
	return "`" + input + "`"
}
func GenerateOtherScripts(config []byte, path string) bool {
	tmpl, err := template.New("dotenv").Parse(_DOTENV)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	if ifExists(dataMapContainer, "fabricVersion") {
		version, _ := dataMapContainer["fabricVersion"].(string)
		core, thridParty := GetVersions(version)
		dataMapContainer["fabricVersion"] = core
		dataMapContainer["thirdPartyVersion"] = thridParty

	} else {
		dataMapContainer["fabricVersion"] = "1.1.0"
		dataMapContainer["thirdPartyVersion"] = "1.0.6"
	}

	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the .env file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+".env", outputBytes.Bytes(), 0666)
	tmpl, err = template.New("setenv").Parse(_SetEnv)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	var outputBytes2 bytes.Buffer
	err = tmpl.Execute(&outputBytes2, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the setenv.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+"setenv.sh", outputBytes2.Bytes(), 0777)
	tmpl, err = template.New("cleanup").Parse(_CleanUp)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	var outputBytes3 bytes.Buffer

	err = tmpl.Execute(&outputBytes3, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the cleanup.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+"cleanup.sh", outputBytes3.Bytes(), 0777)

	tmpl, err = template.New("download").Parse(_DOWNLOAD_SCRIPTS)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	var outputBytes4 bytes.Buffer
	err = tmpl.Execute(&outputBytes4, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the downloadbin.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+"downloadbin.sh", outputBytes4.Bytes(), 0777)

	return true
}
func GenerateGenerateArtifactsScript(config []byte, filename string) bool {
	funcMap := template.FuncMap{
		"ToCMDString": ToCMDString,
		"ToLower":     strings.ToLower,
		"ToUpper":     strings.ToUpper,
	}
	templateToUse := _GenerateArtifactsTemplate
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	addCA := getBoolean(dataMapContainer["addCA"])
	if addCA == true {
		templateToUse = _GenerateArtifactsTemplateWithCA
	}
	tmpl, err := template.New("generateArtifacts").Funcs(funcMap).Parse(templateToUse)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the generateArtifacts file %v\n", err)
		return false
	}
	ioutil.WriteFile(filename, outputBytes.Bytes(), 0777)
	return true
}
func GenerateSetPeer(config []byte, filename string) bool {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("setPeer").Funcs(funcMap).Parse(_SetPeerTemplate)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the setpeer.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(filename, outputBytes.Bytes(), 0777)
	return true
}
func GenerateBuildAndJoinChannelScript(config []byte, filename string) bool {
	funcMap := template.FuncMap{
		"ToCMDString": ToCMDString,
		"ToLower":     strings.ToLower,
	}

	tmpl, err := template.New("buildChannel").Funcs(funcMap).Parse(_BuildChannelScript)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}
	dataMapContainer := make(map[string]interface{})
	channelMap := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	orgs, _ := dataMapContainer["orgs"].([]interface{})
	for _, org := range orgs {
		orgConfig := getMap(org)
		peerCount := getNumber(orgConfig["peerCount"])
		peerNames := make([]string, 0)
		fmt.Printf(" Peer count %d\n", peerCount)
		for index := 0; index < peerCount; index++ {
			peerNames = append(peerNames, fmt.Sprintf("peer%d", index))
		}
		orgConfig["peerNames"] = peerNames
		orgName := getString(orgConfig["name"])
		channelMap[orgName] = orgConfig
	}
	channelMap["channels"] = dataMapContainer["channels"]
	//Resolve the orderer name
	ordererConfig := getMap(dataMapContainer["orderers"])
	if ifExists(ordererConfig, "type") && ifExists(ordererConfig, "haCount") && getString(ordererConfig["type"]) == "kafka" {
		channelMap["ordererURL"] = fmt.Sprintf("%s0.%s:7050", getString(ordererConfig["ordererHostname"]), getString(ordererConfig["domain"]))
	} else {
		channelMap["ordererURL"] = fmt.Sprintf("%s.%s:7050", getString(ordererConfig["ordererHostname"]), getString(ordererConfig["domain"]))
	}
	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, channelMap)
	if err != nil {
		fmt.Printf("Error in generating the buildchannel.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(filename, outputBytes.Bytes(), 0777)
	return true
}
func GetVersions(version string) (string, string) {
	tmpl, err := template.New("versionMap").Parse(_VERSION_COMP_MAP)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return "1.0.0", "1.0.0"
	}
	dataMapContainer := make(map[string]interface{})
	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the version map file %v\n", err)
		return "1.0.0", "1.0.0"
	}
	versionMap := make(map[string]map[string]string)
	json.Unmarshal(outputBytes.Bytes(), &versionMap)
	if _, isOk := versionMap[version]; !isOk {
		fmt.Println("Invalid version number provided defaulting to 1.0.0")
		return "1.0.0", "1.0.0"
	}
	coreVersion := versionMap[version]["fabricCore"]
	thirdPartyVersion := versionMap[version]["thirdParty"]
	return coreVersion, thirdPartyVersion
}
