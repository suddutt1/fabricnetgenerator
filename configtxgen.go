package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
)

const _OrdererTemplate = `
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
    {{.channelName}}Channel:
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
    
Application: &ApplicationDefaults
    Organizations:
`

func GenerateConfigTxGen(config []byte, filename string) bool {

	tmpl, err := template.New("configtxsolo").Parse(_OrdererTemplate)
	if err != nil {
		fmt.Printf("Error in reading template %v\n", err)
		return false
	}
	dataMapContainer := make(map[string]interface{})
	json.Unmarshal(config, &dataMapContainer)
	var outputBytes bytes.Buffer
	err = tmpl.Execute(&outputBytes, dataMapContainer)
	if err != nil {
		fmt.Printf("Error in generating the configtx.yaml file %v\n", err)
		return false
	}
	ioutil.WriteFile(filename, outputBytes.Bytes(), 0666)
	return true
}
