package main

import "io/ioutil"

//GenerateExampleConfig Generate exmaple config
func GenerateExampleConfig(version, basePath string) {
	ioutil.WriteFile(basePath+"network-config.json", []byte(_network_config_v1_example), 0666)
}

const _network_config_v1_example = `
{
    "fabricVersion":"2.2.0",
    "orderers":{
        "name" :"Orderer",
        "mspID":"OrdererMSP",
        "domain":"supplychain.net",
        "ordererHostname":"orderer",
        "SANS":"localhost",
        "caCountry": "IN",
        "caProvince": "Delhi-NCR",
        "caLocality": "Delhi",
        "caOrganizationalUnit": "IT",
        "caStreetAddress": "1, M.G ROAD",
        "caPostalCode": "100001",
        "type":"raft",
        "haCount":3
    },
    "addCA":"true",
    "orgs":[
        { 
            "name" :"Buyer",
            "domain":"superbuyer.com",
            "mspID":"BuyerMSP",
            "SANS":"localhost",
            "caCountry":"IN",
            "caProvince":"Delhi-NCR",
            "caLocality":"Delhi",
            "caOrganizationalUnit":"IT",
            "caStreetAddress":"1, M.G ROAD",
            "caPostalCode":"100001",
            "peerCount":2,
            "userCount":2
        },
        { 
            "name" :"Seller",
            "domain":"rapidseller.net",
            "mspID":"SellerMSP",
            "SANS":"localhost",
            "caCountry":"IN",
            "caProvince":"West Bengal",
            "caLocality":"Kolkata",
            "caOrganizationalUnit":"IT",
            "caStreetAddress":"1, M.G ROAD",
            "caPostalCode":"700001",
            "peerCount":2,
            "userCount":2
        },
        { 
            "name" :"Transporter",
            "domain":"transnet.com",
            "mspID":"TransporterMSP",
            "SANS":"localhost",
            "caCountry":"IN",
            "caProvince":"Karnataka",
            "caLocality":"Bangaluru",
            "caOrganizationalUnit":"IT",
            "caStreetAddress":"1, M.G ROAD",
            "caPostalCode":"560001",
            "peerCount":2,
            "userCount":2
        }
        ],
    "extraHosts":{
        "myhost":"127.0.0.1"
    },
    "consortium":"SupplyChainConsortium",
    "channels" :[
                    {"channelName":"Sales","orgs":["Buyer","Seller"] },
                    {"channelName":"Logistics","orgs":["Buyer","Seller","Transporter"] }
                ],
    "chaincodes":[{"channelName":"Sales","ccid":"salestrade","version":"1.0","src":"github.com/salestrade","participants":["Buyer","Seller"]}]            
                
}


`
