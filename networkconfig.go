package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

//NetworkConfig represents the network configuration
type NetworkConfig struct {
	Version     string                `json:"fabricVersion"`
	Consortinum string                `json:"consortium"`
	Orgs        []OrganizationDetails `json:"orgs"`
	Orderer     OrdererDetails        `json:"orderers"`
	Channels    []ChannelDetails      `json:"channels"`
	Chaincodes  []ChaincodeDetails    `json:"chaincodes"`
}

//OrganizationDetails contains organization details
type OrganizationDetails struct {
	Name   string `json:"name"`
	MSPID  string `json:"mspID"`
	Domain string `json:"domain"`
	SANS   string `json:"SANS"`
	Peers  int    `json:"peerCount"`
	Users  int    `json:"userCount"`
}

//OrdererDetails contains orders details
type OrdererDetails struct {
	Name            string `json:"name"`
	MSPID           string `json:"mspID"`
	Domain          string `json:"domain"`
	SANS            string `json:"SANS"`
	OrdererHostname string `json:"ordererHostname"`
	Type            string `json:"type"`
	HACount         int    `json:"haCount"`
}

//ChannelDetails represents the fabric channel details
type ChannelDetails struct {
	Name string   `json:"channelName"`
	Orgs []string `json:"orgs"`
}

//ChaincodeDetails represent chaincode details
type ChaincodeDetails struct {
	Channel      string   `json:"channelName"`
	CCID         string   `json:"ccid"`
	Version      string   `json:"version"`
	SrcPath      string   `json:"src"`
	Participants []string `json:"participants"`
}

//NewNetworkConfig returns a new network config
func NewNetworkConfig(filePath string) NetworkConfig {
	var nc NetworkConfig
	configBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Unable to open the file %+v", err)
	}
	err = json.Unmarshal(configBytes, &nc)
	if err != nil {
		log.Fatalf("Unable to parse the config %+v", err)
	}
	return nc
}

//BuildPeerConnectionProfile builds the peer connection profile snippet
func (od *OrganizationDetails) BuildPeerConnectionProfile() map[string]interface{} {
	profile := make(map[string]interface{})
	for peerIndex := 0; peerIndex < od.Peers; peerIndex++ {
		peerFQDN := fmt.Sprintf("peer%d.%s", peerIndex, od.Domain)
		grpcOpts := map[string]interface{}{
			"ssl-target-name-override": peerFQDN,
			"keep-alive-time":          "20s",
			"keep-alive-timeout":       "100s",
			"keep-alive-permit":        false,
			"fail-fast":                false,
			"allow-insecure":           false,
		}
		tlsCACerts := map[string]interface{}{
			"path": fmt.Sprintf("/networkPath/crypto-config/peerOrganizations/%s/tlsca/tlsca.%s-cert.pem", od.Domain, od.Domain),
		}
		profile[peerFQDN] = map[string]interface{}{
			"grpcOptions": grpcOpts,
			"tlsCACerts":  tlsCACerts,
		}

	}
	return profile
}

//BuildCAEntries returns certificate authority enties
func (od *OrganizationDetails) BuildCAEntries() interface{} {
	caEntry := make(map[string]interface{})
	caDetails := struct {
		URL        string            `yaml:"url"`
		HTTPS      map[string]bool   `yaml:"httpsOptions"`
		TLSCACert  map[string]string `yaml:"tlsCACerts"`
		CANAme     string            `yaml:"caName"`
		Registerer map[string]string `yaml:"registrar"`
	}{
		URL: fmt.Sprintf("ca.%s:7054", od.Domain),
		HTTPS: map[string]bool{
			"verify": false,
		},
		TLSCACert: map[string]string{
			"path": fmt.Sprintf("/networkPath/crypto-config/peerOrganizations/%s/ca/ca.%s-cert.pem", od.Domain, od.Domain),
		},
		CANAme: fmt.Sprintf("%sCA", od.Name),
		Registerer: map[string]string{
			"enrollId":     "admin",
			"enrollSecret": "adminpw",
		},
	}
	caEntry[fmt.Sprintf("%s-ca", strings.ToLower(od.Name))] = caDetails
	return caEntry
}

//BuildOrgDetails returns the organization level entries
func (od *OrganizationDetails) BuildOrgDetails() interface{} {
	peerList := make([]string, 0)
	for peerIndex := 0; peerIndex < od.Peers; peerIndex++ {
		peerFQDN := fmt.Sprintf("peer%d.%s", peerIndex, od.Domain)
		peerList = append(peerList, peerFQDN)
	}
	orgDetails := struct {
		MSPID           string   `yaml:"mspid"`
		CryptoPath      string   `yaml:"cryptoPath"`
		CertAuthorities []string `yaml:"certificateAuthorities"`
		Peers           []string `yaml:"peers"`
	}{
		MSPID:           od.MSPID,
		CryptoPath:      fmt.Sprintf("peerOrganizations/%s/users/{username}@%s/msp", od.Domain, od.Domain),
		CertAuthorities: []string{fmt.Sprintf("%s-ca", strings.ToLower(od.Name))},
		Peers:           peerList,
	}
	return orgDetails
}

//BuildChannelDetails returns the channel entries
func (od *OrganizationDetails) BuildChannelDetails(orderers []string) interface{} {
	peerEntry := map[string]interface{}{
		"endorsingPeer":  true,
		"chaincodeQuery": true,
		"ledgerQuery":    true,
		"eventSource":    true,
	}
	pols := map[string]interface{}{
		"queryChannelConfig": map[string]interface{}{
			"minResponses": 1,
			"maxTargets":   1,
			"retryOpts": map[string]interface{}{
				"attempts":       "2",
				"initialBackoff": "500ms",
				"maxBackoff":     "5s",
				"backoffFactor":  1.0,
			},
		},
		"discovery": map[string]interface{}{
			"maxTargets": 1,
			"retryOpts": map[string]interface{}{
				"attempts":       "2",
				"initialBackoff": "500ms",
				"maxBackoff":     "5s",
				"backoffFactor":  2.0,
			},
		},
		"eventService": map[string]interface{}{
			"resolverStrategy":                 "PreferOrg",
			"balancer":                         "Random",
			"blockHeightLagThreshold":          5,
			"reconnectBlockHeightLagThreshold": 8,
			"peerMonitorPeriod":                "6s",
		},
	}
	peers := make(map[string]interface{})
	for peerIndex := 0; peerIndex < od.Peers; peerIndex++ {
		peerFQDN := fmt.Sprintf("peer%d.%s", peerIndex, od.Domain)
		peers[peerFQDN] = peerEntry
	}
	channelEntry := struct {
		Orderer  []string               `yaml:"orderers"`
		Peers    map[string]interface{} `yaml:"peers"`
		Policies map[string]interface{} `yaml:"policies"`
	}{
		Orderer:  orderers,
		Peers:    peers,
		Policies: pols,
	}
	return channelEntry
}

//BuildClientEntry builds the client entry for connection profile
func (od *OrganizationDetails) BuildClientEntry() interface{} {
	cpMap := struct {
		Client  string                 `yaml:"client"`
		Logging map[string]interface{} `yaml:"logging"`
		Peer    map[string]interface{} `yaml:"peer"`
		Orderer map[string]interface{} `yaml:"orderer"`
		Global  map[string]interface{} `yaml:"global"`
	}{
		Client: strings.ToLower(od.Name),
		Logging: map[string]interface{}{
			"level": "debug",
		},
		Peer: map[string]interface{}{
			"timeout": map[string]interface{}{
				"connection": "100s",
				"response":   "600s",
				"discovery": map[string]interface{}{
					"greylistExpiry": "100s",
				},
			},
		},
		Orderer: map[string]interface{}{
			"timeout": map[string]interface{}{
				"connection": "100s",
				"response":   "600s",
			},
		},
		Global: map[string]interface{}{
			"timeout": map[string]interface{}{
				"query":   "180s",
				"execute": "180s",
				"resmgmt": "180s",
			},
		},
	}
	return cpMap
}

//GenerateConnectionProfile generates all the organization connection profiles
func (nc *NetworkConfig) GenerateConnectionProfile() {
	ordererList := nc.Orderer.GetFQDNList()
	//For each one the organizaition generate connection profile
	for _, org := range nc.Orgs {
		//Get the channels those are part of this organization
		channelList := make([]string, 0)
		for _, chDetails := range nc.Channels {
			for _, orgName := range chDetails.Orgs {
				if orgName == org.Name {
					channelList = append(channelList, chDetails.Name)
					break
				}
			}
		}
		//For all the channels build the channel entries
		channelMap := make(map[string]interface{})
		for _, chName := range channelList {
			channelMap[chName] = org.BuildChannelDetails(ordererList)
		}
		finalConfig := struct {
			Ver        string      `yaml:"version"`
			Client     interface{} `yaml:"client"`
			ChannelMap interface{} `yaml:"channels"`
			Orderer    interface{} `yaml:"orderers"`
			Peers      interface{} `yaml:"peers"`
			Orgs       interface{} `yaml:"organizations"`
			CAEntries  interface{} `yaml:"certificateAuthorities"`
		}{
			Ver:        "1.0.0",
			Client:     org.BuildClientEntry(),
			ChannelMap: channelMap,
			Orderer:    nc.Orderer.BuildOSNDetails(),
			Peers:      org.BuildPeerConnectionProfile(),
			Orgs: map[string]interface{}{
				strings.ToLower(org.Name): org.BuildOrgDetails(),
				"ordererorg":              nc.Orderer.BuildOrdererOrgDetails(),
			},
			CAEntries: org.BuildCAEntries(),
		}

		output, err := yaml.Marshal(finalConfig)
		if err != nil {
			fmt.Printf("Error in marshalling %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Output \n%s\n", string(output))
	}

}

//GetFQDNList returns the all orderer instance's FQDN list
func (od *OrdererDetails) GetFQDNList() []string {
	ordererHostNames := make([]string, 0)
	if od.HACount > 0 {
		for haIndex := 0; haIndex < od.HACount; haIndex++ {
			ordererHostNames = append(ordererHostNames, fmt.Sprintf("%s%d.%s", od.OrdererHostname, haIndex, od.Domain))
		}
	} else {
		ordererHostNames = append(ordererHostNames, fmt.Sprintf("%s.%s", od.OrdererHostname, od.Domain))
	}
	return ordererHostNames
}

//BuildOrdererOrgDetails retruns orderer organization level entries
func (od *OrdererDetails) BuildOrdererOrgDetails() interface{} {
	ordererDetails := struct {
		MSPID      string `yaml:"mspID"`
		CryptoPath string `yaml:"cryptoPath"`
	}{
		MSPID:      od.MSPID,
		CryptoPath: fmt.Sprintf("ordererOrganizations/%s/users/{username}@%s/msp", od.Domain, od.Domain),
	}
	return ordererDetails
}

//BuildOSNDetails returns the ordering service nodes
func (od *OrdererDetails) BuildOSNDetails() interface{} {
	osnMap := make(map[string]interface{})
	if od.HACount > 0 {
		for ordIndex := 0; ordIndex < od.HACount; ordIndex++ {
			fqdn := fmt.Sprintf("%s%d.%s", od.OrdererHostname, ordIndex, od.Domain)
			ordererDetails := struct {
				URL       string                 `yaml:"url"`
				GRPC      map[string]interface{} `yaml:"grpcOptions"`
				TLSCACert map[string]interface{} `yaml:"tlsCACerts"`
			}{
				URL: fmt.Sprintf("%s:7050", fqdn),
				GRPC: map[string]interface{}{
					"ssl-target-name-override": fqdn,
					"keep-alive-time":          "20s",
					"keep-alive-timeout":       "100s",
					"keep-alive-permit":        false,
					"fail-fast":                false,
					"allow-insecure":           false,
				},
				TLSCACert: map[string]interface{}{
					"path": fmt.Sprintf("ordererOrganizations/%s/users/{username}@%s/msp", od.Domain, od.Domain),
				},
			}

			osnMap[fqdn] = ordererDetails
		}
	} else {
		fqdn := fmt.Sprintf("%s.%s", od.OrdererHostname, od.Domain)
		ordererDetails := struct {
			URL       string                 `yaml:"url"`
			GRPC      map[string]interface{} `yaml:"grpcOptions"`
			TLSCACert map[string]interface{} `yaml:"tlsCACerts"`
		}{
			URL: fmt.Sprintf("%s:7050", fqdn),
			GRPC: map[string]interface{}{
				"ssl-target-name-override": fqdn,
				"keep-alive-time":          "20s",
				"keep-alive-timeout":       "100s",
				"keep-alive-permit":        false,
				"fail-fast":                false,
				"allow-insecure":           false,
			},
			TLSCACert: map[string]interface{}{
				"path": fmt.Sprintf("ordererOrganizations/%s/users/{username}@%s/msp", od.Domain, od.Domain),
			},
		}

		osnMap[fqdn] = ordererDetails

	}

	return osnMap
}