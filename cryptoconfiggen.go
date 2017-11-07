package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func GenerateCrytoConfig(config []byte, filePath string) bool {
	isSuccess := true
	rootConfig := make(map[string]interface{})
	err := json.Unmarshal(config, &rootConfig)
	if err != nil {
		fmt.Printf("Error in parsing the config bytes %v", err)
		return false
	}
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
		peerOrgs = append(peerOrgs, buildOrgConfig(getMap(orgConfig)))
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
	return outputStructure
}
func buildOrgConfig(orgConfig map[string]interface{}) map[string]interface{} {
	outputStructure := make(map[string]interface{})
	outputStructure["Name"] = getString(orgConfig["name"])
	outputStructure["Domain"] = getString(orgConfig["domain"])
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
func getMap(element interface{}) map[string]interface{} {
	retMap, ok := element.(map[string]interface{})
	if ok == true {
		return retMap
	}
	return nil
}
func getString(element interface{}) string {
	retString, ok := element.(string)
	if ok == true {
		return retString
	}
	return ""
}
func getNumber(element interface{}) int {

	s := fmt.Sprintf("%v", element)
	retString, err := strconv.Atoi(s)
	if err == nil {
		return retString
	}
	return 0
}
