package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

//GenerateCrytoConfig generate the CryptoConfig file
func GenerateCrytoConfig(config []byte, filePath string) bool {
	useCA := false
	isSuccess := true
	rootConfig := make(map[string]interface{})
	err := json.Unmarshal(config, &rootConfig)
	if err != nil {
		fmt.Printf("Error in parsing the config bytes %v", err)
		return false
	}
	useCA = getBoolean(rootConfig["addCA"])
	//Perform the orderer part
	ordererConfig := getMap(rootConfig["orderers"])
	if ordererConfig == nil {
		fmt.Println("No orderer specified")
		return false
	}

	cryptoConfig := make(map[string]interface{})
	orderOrgs := make([]map[string]interface{}, 0)
	orderOrgs = append(orderOrgs, buildOrderConfig(ordererConfig))
	cryptoConfig["OrdererOrgs"] = orderOrgs
	orgs, orgsExists := rootConfig["orgs"].([]interface{})
	if !orgsExists {
		fmt.Println("No organizations specified")
		return false
	}
	peerOrgs := make([]map[string]interface{}, 0)
	for _, orgConfig := range orgs {
		peerOrgs = append(peerOrgs, buildOrgConfig(getMap(orgConfig), useCA))
	}
	cryptoConfig["PeerOrgs"] = peerOrgs
	outBytes, _ := yaml.Marshal(cryptoConfig)
	//fmt.Printf("Crypto config Orderes\n%s\n", string(outBytes))

	ioutil.WriteFile(filePath, outBytes, 0666)
	return isSuccess
}
func buildOrderConfig(ordererConfig map[string]interface{}) map[string]interface{} {
	outputStructure := make(map[string]interface{})
	outputStructure["Name"] = getString(ordererConfig["name"])
	outputStructure["Domain"] = getString(ordererConfig["domain"])
	specs := make([]map[string]interface{}, 0)
	hostnameSpec := make(map[string]interface{})
	hostnameSpec["Hostname"] = getString(ordererConfig["ordererHostname"])
	specs = append(specs, hostnameSpec)
	//Assuing one as of now
	sansInput := strings.Split(getString(ordererConfig["SANS"]), ",")

	sansArray := make([]string, len(sansInput))
	for indx, sans := range sansInput {
		sansArray[indx] = sans
	}
	sansSpec := make(map[string]interface{})
	sansSpec["SANS"] = sansArray
	specs = append(specs, sansSpec)
	outputStructure["Specs"] = specs
	if ifExists(ordererConfig, "haCount") && ifExists(ordererConfig, "type") {
		if getString(ordererConfig["type"]) == "kafka" {
			template := make(map[string]interface{})
			template["Count"] = ordererConfig["haCount"]
			outputStructure["Template"] = template
		}
	}

	return outputStructure
}
func buildOrgConfig(orgConfig map[string]interface{}, useCA bool) map[string]interface{} {
	outputStructure := make(map[string]interface{})
	outputStructure["Name"] = getString(orgConfig["name"])
	outputStructure["Domain"] = getString(orgConfig["domain"])
	if useCA == true {
		caTemplate := make(map[string]string)
		caTemplate["Hostname"] = "ca"
		outputStructure["CA"] = caTemplate
	}

	template := make(map[string]interface{})
	template["Count"] = orgConfig["peerCount"]
	//Assuing one as of now
	sansArray := make([]string, 1)
	sansArray[0] = getString(orgConfig["SANS"])
	template["SANS"] = sansArray
	outputStructure["Template"] = template
	users := make(map[string]interface{})
	users["Count"] = orgConfig["userCount"]
	outputStructure["Users"] = users
	return outputStructure
}
