package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
)

func GenerateConfigTxGen(config []byte, filename string) bool {

	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	configTxTemplate := _configTxTemplateV13
	confVersion := getString(dataMapContainer["fabricVersion"])
	if confVersion == "1.4.2" {
		fmt.Println("Generation 1.4.x compatible configtxgen")
		configTxTemplate = _configTxTemplateV142Raft
	} else {
		fmt.Println("Generation 1.4.0 compatible configtxgen")
	}
	tmpl, err := template.New("configtxsolo").Parse(configTxTemplate)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}

	ordererConfig := getMap(dataMapContainer["orderers"])
	if ifExists(ordererConfig, "type") && ifExists(ordererConfig, "haCount") {
		orderType := getString(ordererConfig["type"])
		if orderType == "kafka" || orderType == "raft" {
			hostName := getString(ordererConfig["ordererHostname"])
			domainName := getString(ordererConfig["domain"])
			listOfOrderers := make([]string, 0)
			listOfOrdererConcenters := make([]map[string]string, 0)

			//Quick fix for orderer port sequence
			port := 7050
			for index := 0; index < getNumber(ordererConfig["haCount"]); index++ {
				listOfOrderers = append(listOfOrderers, fmt.Sprintf("%s%d.%s:%d", hostName, index, domainName, 7050))
				port += 1000
				cententerEntry := map[string]string{
					"hostname":      fmt.Sprintf("%s%d.%s", hostName, index, domainName),
					"port":          "7050",
					"clientTLSCert": fmt.Sprintf("crypto-config/ordererOrganizations/%s/orderers/%s%d.%s/tls/server.crt", domainName, hostName, index, domainName),
					"serverTLSCert": fmt.Sprintf("crypto-config/ordererOrganizations/%s/orderers/%s%d.%s/tls/server.crt", domainName, hostName, index, domainName),
				}
				listOfOrdererConcenters = append(listOfOrdererConcenters, cententerEntry)
			}
			dataMapContainer["ordererFDQNList"] = listOfOrderers
			dataMapContainer["consenters"] = listOfOrdererConcenters
		}
	}

	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the configtx.yaml file %v\n", err)
		return false
	}
	err = ioutil.WriteFile(filename, outputBytes.Bytes(), 0666)
	if err != nil {
		fmt.Printf("Error in generating file %v\n", err)
		return false
	}
	return true
}
