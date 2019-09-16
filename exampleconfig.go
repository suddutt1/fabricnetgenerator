package main

import "io/ioutil"

//GenerateExampleConfig Generate exmaple config
func GenerateExampleConfig(version, basePath string) {
	ioutil.WriteFile(basePath+"network-config.json", []byte(_network_config_v1_example), 0666)
}

const _network_config_v1_example = `
{
    "fabricVersion":"1.4.2",
    "orderers":{
        "name" :"Orderer","mspID":"OrdererMSP","domain":"supplychain.net","ordererHostname":"orderer","SANS":"localhost","type":"raft","haCount":3
    },
    "addCA":"true",
    "orgs":[
        { 
            "name" :"Buyer",
            "domain":"superbuyer.com",
            "mspID":"BuyerMSP",
            "SANS":"localhost",
            "peerCount":2,
            "userCount":2
        },
        { 
            "name" :"Seller",
            "domain":"rapidseller.net",
            "mspID":"SellerMSP",
            "SANS":"localhost",
            "peerCount":2,
            "userCount":2
        },
        { 
            "name" :"Transporter",
            "domain":"transnet.com",
            "mspID":"TransporterMSP",
            "SANS":"localhost",
            "peerCount":2,
            "userCount":2
        }
        ],
    "consortium":"SupplyChainConsortium",
    "channels" :[
                    {"channelName":"Sales","orgs":["Buyer","Seller"] },
                    {"channelName":"Logistics","orgs":["Buyer","Seller","Transporter"] }
                ],
    "chaincodes":[{"channelName":"Sales","ccid":"salestrade","version":"1.0","src":"github.com/salestrade","participants":["Buyer","Seller"]}]            
                
}


`
