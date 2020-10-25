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

#For fabric 2.2.x extra environment variables
{{range .orgs}}
export {{.name | ToUpper }}_PEER0_CA=/opt/ws/crypto-config/peerOrganizations/{{.domain}}/peers/peer0.{{.domain}}/tls/ca.crt
{{end}}

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
	
		$CONFIGTXGEN -profile OrdererGenesis -channelID system-channel -outputBlock ./genesis.block  -channelID genesischannel
		
		{{range .channels}}{{$chName := .channelName }}{{$channelId:= $chName | ToLower }}
		$CONFIGTXGEN -profile {{print $chName }} -outputCreateChannelTx ./{{print $channelId ".tx" }} -channelID {{ print $channelId }}
		{{range $org:= .orgs}}
		echo {{print "\"Generating anchor peers tx files for  " $org "\""}}
		$CONFIGTXGEN -profile {{print $chName }} -outputAnchorPeersUpdate  ./{{print $channelId  $org "MSPAnchor.tx" }} -channelID {{ print $channelId }} -asOrg {{print $org "MSP" }}
		{{end}}
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
	
		$CONFIGTXGEN -profile OrdererGenesis -channelID system-channel -outputBlock ./genesis.block
		{{range .channels}}{{$chName := .channelName }}{{$channelId:= $chName | ToLower }}
		$CONFIGTXGEN -profile {{print $chName }} -outputCreateChannelTx ./{{print $channelId ".tx" }} -channelID {{ print $channelId }}
		{{range $org:= .orgs}}
		echo {{print "\"Generating anchor peers tx files for  " $org "\""}}
		$CONFIGTXGEN -profile {{print $chName }} -outputAnchorPeersUpdate  ./{{print $channelId  $org "MSPAnchor.tx" }} -channelID {{ print $channelId }} -asOrg {{print $org "MSP" }}
		{{end}}

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
{{ $timeOut:= .timeout}}
{{ $root := . }}
{{range .channels}}
	{{ $channelId := print .channelName  | ToLower }}
	echo "Building channel for {{print $channelId}}" 
	{{$firstOrg := (index .orgs 0) }}
	. setpeer.sh {{$firstOrg}} peer0
	export CHANNEL_NAME="{{print $channelId }}"
	peer channel create -o {{ print $orderer }} -c $CHANNEL_NAME -f ./{{print $channelId ".tx"}} --tls true --cafile $ORDERER_CA -t {{$timeOut}}
	{{ range $index,$orgName :=.orgs}}
		{{$orgConfig :=  index $root $orgName }}
        {{ range $i,$peerId:=$orgConfig.peerNames }}
            . setpeer.sh {{$orgName}} {{$peerId}}
            export CHANNEL_NAME="{{print $channelId }}"
			peer channel join -b $CHANNEL_NAME.block
		{{end}}
	{{end}}
	{{ range $index,$orgName :=.orgs}}
	{{$orgConfig :=  index $root $orgName }}
	{{ range $i,$peerId:=$orgConfig.peerNames }}
		. setpeer.sh {{$orgName}} {{$peerId}}
		export CHANNEL_NAME="{{print $channelId }}"
		peer channel update -o  {{ print $orderer }} -c $CHANNEL_NAME -f ./{{print $channelId  $orgName "MSPAnchor.tx" }} --tls --cafile $ORDERER_CA 
	{{end}}
{{end}}

{{end}}

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
rm add_affi*.sh
rm -rf *.log
rm -rf *log.txt
rm -rf *.tar.gz

`
const _SetEnv_1_4_0 = `
#!/bin/bash
export IMAGE_TAG="1.4.0"
export COUCH_TAG="0.4.15"
export ZK_TAG="0.4.14"
export KAFKA_TAG="0.4.14"
export TOOLS_TAG="1.4.0"
export TAG_CCENV="1.4.0"
export TAG_BASEOS="0.4.14"
export CA_TAG="1.4.0"

`
const _SetEnv_1_4_2 = `
#!/bin/bash
export IMAGE_TAG="1.4.2"
export TOOLS_TAG="1.4.2"
export TAG_CCENV="1.4.2"
export COUCH_TAG="0.4.15"
export TAG_BASEOS="0.4.15"
export CA_TAG="1.4.2"

export KAFKA_TAG="0.4.15"
export ZK_TAG="0.4.15"

`
const _SetEnv_2_2 = `
#!/bin/bash
export IMAGE_TAG="2.2.0"
export TOOLS_TAG="2.2.0"
export TAG_CCENV="2.2.0"
export COUCH_TAG="3.1"
export TAG_BASEOS="2.2.0"
export CA_TAG="1.4.7"


`
const _GitIgnore = `
worldstate/
*.tar.gz
*.txt
*.block
*.tx
bin/
blocks/
crypto-config/*
config/
bin/
chaincode/*


`

const _DOTENV = `
COMPOSE_PROJECT_NAME=bc

`
const _DOWNLOAD_SCRIPTS = `
#!/bin/bash

export VERSION={{.downloadVersion}}
export ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
echo "===> Downloading platform binaries"
export URL="https://github.com/hyperledger/fabric/releases/download/v${VERSION}/hyperledger-fabric-${ARCH}-${VERSION}.tar.gz"
echo $URL
curl  -L $URL| tar xz

`
const _VERSION_COMP_MAP = `
{
	"1.4.0":{ "fabricCore":"1.4.0","thirdParty":"0.4.14","couch":"0.4.15","zk":"0.4.14","kafka":"0.4.14"}
}	

`
const _REMOVE_IMAGES = `
#!/bin/bash
docker rmi -f {{ "docker images bc* -aq" | ToCMDString}}
`
const _CREATE_AFFILIATION = `
#!/bin/bash
fabric-ca-client enroll  -u https://admin:adminpw@ca.{{.domain}}:7054 --tls.certfiles /etc/hyperledger/fabric-ca-server-config/ca.{{.domain}}-cert.pem 
fabric-ca-client affiliation add {{.orgName | ToLower }}  -u https://admin:adminpw@ca.{{.domain}}:7054 --tls.certfiles /etc/hyperledger/fabric-ca-server-config/ca.{{.domain}}-cert.pem 
`

const _CLEAR_VOL_SCRIPT = `
#!/bin/bash
sudo rm -rf ./worldstate
sudo rm -rf ./blocks

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

	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the .env file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+".env", outputBytes.Bytes(), 0666)
	fabVersion := "1.4.0"
	version, _ := dataMapContainer["fabricVersion"].(string)
	if ifExists(dataMapContainer, "fabricVersion") {
		fabVersion = version
	}
	tmplName := _SetEnv_1_4_0
	switch fabVersion {
	case "1.4.0":
		tmplName = _SetEnv_1_4_0
	case "1.4.2":
		tmplName = _SetEnv_1_4_2
	case "2.2.0":
		tmplName = _SetEnv_2_2
	}

	tmpl, err = template.New("setenv").Parse(tmplName)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}
	var outputBytes2 bytes.Buffer
	err = tmpl.Execute(&outputBytes2, nil)
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
	dataMapContainer["downloadVersion"] = version
	var outputBytes4 bytes.Buffer
	err = tmpl.Execute(&outputBytes4, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the downloadbin.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+"downloadbin.sh", outputBytes4.Bytes(), 0777)
	funcMap := template.FuncMap{
		"ToCMDString": ToCMDString,
		"ToLower":     strings.ToLower,
		"ToUpper":     strings.ToUpper,
	}

	tmpl, err = template.New("rmImages").Funcs(funcMap).Parse(_REMOVE_IMAGES)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	var outputBytes5 bytes.Buffer
	err = tmpl.Execute(&outputBytes5, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the generateArtifacts file %v\n", err)
		return false
	}
	ioutil.WriteFile(path+"removeImages.sh", outputBytes5.Bytes(), 0777)
	ioutil.WriteFile(path+"clearVols.sh", []byte(_CLEAR_VOL_SCRIPT), 0777)
	ioutil.WriteFile(path+".gitignore", []byte(_GitIgnore), 0666)
	GenerateAffiliationScripts(config, path)
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

//GenerateSetPeer emits setpeer.sh file
func GenerateSetPeer(config []byte, filename string) bool {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
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

//GenerateBuildAndJoinChannelScript generates build and join channel script
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
	//Check the version
	fabricVersion, _ := dataMapContainer["fabricVersion"].(string)
	if strings.HasPrefix(fabricVersion, "2.2") {
		return GenerateBuildAndJoinChannelScriptV2x(config)
	}

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
	if ifExists(ordererConfig, "type") && ifExists(ordererConfig, "haCount") && (getString(ordererConfig["type"]) == "kafka" || getString(ordererConfig["type"]) == "raft") {
		channelMap["ordererURL"] = fmt.Sprintf("%s0.%s:7050", getString(ordererConfig["ordererHostname"]), getString(ordererConfig["domain"]))
	} else {
		channelMap["ordererURL"] = fmt.Sprintf("%s.%s:7050", getString(ordererConfig["ordererHostname"]), getString(ordererConfig["domain"]))
	}

	timeOut := "1000"
	if IsVersionAbove(dataMapContainer, "1.3.0") {
		fmt.Println("Generation 1.3 compatible build&joinchannel")
		timeOut = "1000s"
	}
	channelMap["timeout"] = timeOut
	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, channelMap)
	if err != nil {
		fmt.Printf("Error in generating the buildchannel.sh file %v\n", err)
		return false
	}
	ioutil.WriteFile(filename, outputBytes.Bytes(), 0777)
	return true
}
func GetVersions(version string) map[string]string {
	versionMap := make(map[string]map[string]string)
	json.Unmarshal([]byte(_VERSION_COMP_MAP), &versionMap)
	if details, isOk := versionMap[version]; !isOk {
		fmt.Println("Invalid version number provided defaulting to 1.4.0")
		return versionMap["1.4.0"]
	} else {
		return details
	}

}
func IsVersionAbove(config map[string]interface{}, version string) bool {
	confVersion := getString(config["fabricVersion"])

	return len(version) > 0 && strings.Compare(confVersion, version) >= 0
}
func GenerateAffiliationScripts(config []byte, path string) {
	funcMap := template.FuncMap{

		"ToLower": strings.ToLower,
	}
	networkConfig := make(map[string]interface{})
	json.Unmarshal(config, &networkConfig)
	addCA := getBoolean(networkConfig["addCA"])
	if addCA {
		orgConfigs, _ := networkConfig["orgs"].([]interface{})
		tmpl, err := template.New("addAffiliation").Funcs(funcMap).Parse(_CREATE_AFFILIATION)
		if err != nil {
			fmt.Printf("Error in reading template %v\n", err)
			return
		}

		for _, org := range orgConfigs {
			orgConfig, _ := org.(map[string]interface{})
			templateData := make(map[string]string)
			templateData["domain"] = getString(orgConfig["domain"])
			templateData["orgName"] = getString(orgConfig["name"])
			var outputBytes bytes.Buffer
			err = tmpl.Execute(&outputBytes, templateData)
			if err != nil {
				fmt.Printf("Error in generating the generateArtifacts file %v\n", err)
				return
			}
			fileName := fmt.Sprintf("add_affiliation_%s.sh", strings.ToLower(templateData["orgName"]))
			ioutil.WriteFile(path+fileName, outputBytes.Bytes(), 0777)
		}
	}
}
