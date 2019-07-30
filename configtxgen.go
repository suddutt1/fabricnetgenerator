package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
)

const _configTxTemplate = `
Profiles:

    OrdererGenesis:
        Orderer:
            <<: *OrdererDefaults
            Organizations:
                - *OrdererOrg
        Consortiums:
          {{.consortium}}:
             Organizations:
                {{ range .orgs}}- *{{ .name}}Org
                {{end}}
    {{ $x :=.consortium}}
    {{range .channels}}
    {{.channelName}}:
        Consortium: {{$x}}
        Application:
            <<: *ApplicationDefaults
            Organizations:
                {{range $index,$var := .orgs}}- *{{$var}}Org
                {{end}}
    {{end}} 
Organizations:
    - &OrdererOrg
        Name: {{index .orderers "mspID" }}
        ID: {{index .orderers "mspID" }}
        MSPDir: crypto-config/ordererOrganizations/{{ index .orderers "domain" }}/msp
    {{range .orgs}}
    - &{{ .name}}Org
        Name: {{.mspID}}
        ID: {{.mspID}}
        MSPDir: crypto-config/peerOrganizations/{{ .domain  }}/msp
        AnchorPeers:
          - Host: peer0.{{.domain}}
            Port: 7051
        {{ end }}
{{ if  and (eq .orderers.type "kafka")  (  .orderers.haCount ) }}
Orderer: &OrdererDefaults
        OrdererType: kafka
        Addresses:{{ range .ordererFDQNList }}
          - {{.}}{{end}}
        BatchTimeout: 2s
        BatchSize:
          MaxMessageCount: 10
          AbsoluteMaxBytes: 98 MB
          PreferredMaxBytes: 512 KB
        Kafka:
          Brokers:
            - kafka0:9092
            - kafka1:9092
            - kafka2:9092
            - kafka3:9092
        Organizations:
{{else}}
Orderer: &OrdererDefaults
        OrdererType: solo
        Addresses:
          - {{index .orderers "ordererHostname" }}.{{index .orderers "domain"}}:7050
        BatchTimeout: 2s
        BatchSize:
          MaxMessageCount: 10
          AbsoluteMaxBytes: 98 MB
          PreferredMaxBytes: 512 KB
        Kafka:
          Brokers:
            - 127.0.0.1:9092
        Organizations:

{{end}}    
Application: &ApplicationDefaults
    Organizations:
`

func GenerateConfigTxGen(config []byte, filename string) bool {

	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	configTxTemplate := _configTxTemplateV13
	if IsVersionAbove(dataMapContainer, "1.4.0") {
		fmt.Println("Generation 1.4.x compatible configtxgen")
		configTxTemplate = _configTxTemplateV142Raft
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
