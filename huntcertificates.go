package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

//HuntCertificates hunts the certificates of the users and private keys
func HuntCertificates(baseDir string, isJSON bool) {
	fmt.Println("Hunting user certificates at ", baseDir)
	orgMainPath := baseDir + "/crypto-config/peerOrganizations"
	orgs, err := ioutil.ReadDir(baseDir + "/crypto-config/peerOrganizations")
	if err != nil {
		fmt.Printf("\nError is reading %s %+v", orgMainPath, err)
		return
	}
	for _, orgDirectory := range orgs {
		if orgDirectory.IsDir() {
			outputFileName := fmt.Sprintf("./%s_users.yaml", orgDirectory.Name())
			if isJSON {
				outputFileName = fmt.Sprintf("./%s_users.json", orgDirectory.Name())
			}

			usersEntry := make(map[string]interface{})
			fmt.Println("Going to get the users of the org ", orgDirectory.Name())
			userDir := fmt.Sprintf("%s/%s/users", orgMainPath, orgDirectory.Name())

			userDirs, err := ioutil.ReadDir(userDir)
			if err != nil {
				fmt.Printf("\nError is reading %s %+v", userDir, err)
				return
			}
			for _, userDir := range userDirs {
				fmt.Println("Going to get the certificates of ", userDir.Name())
				certPath := fmt.Sprintf("%s/%s/users/%s/msp/signcerts", orgMainPath, orgDirectory.Name(), userDir.Name())
				certFiles, err := ioutil.ReadDir(certPath)
				if err != nil {
					fmt.Printf("\nError is reading %s %+v", certPath, err)
					return
				}
				certFileName := ""
				for _, certFile := range certFiles {
					if !certFile.IsDir() {
						fmt.Println("Reading file ", certFile.Name())
						certFileName = fmt.Sprintf("%s/%s", certPath, certFile.Name())
					}
				}
				keyPath := fmt.Sprintf("%s/%s/users/%s/msp/keystore", orgMainPath, orgDirectory.Name(), userDir.Name())
				keyFiles, err := ioutil.ReadDir(keyPath)
				if err != nil {
					fmt.Printf("\nError is reading %s %+v", keyPath, err)
					return
				}
				keyFileName := ""
				for _, keyFile := range keyFiles {
					if !keyFile.IsDir() {
						fmt.Println("Reading file ", keyFile.Name())
						keyFileName = fmt.Sprintf("%s/%s", keyPath, keyFile.Name())
					}
				}

				GenerateCertKeyEntry(certFileName, keyFileName, userDir.Name(), isJSON, usersEntry)

			}
			finalEntry := make(map[string]interface{})
			finalEntry["users"] = usersEntry
			if isJSON {
				finalOutput, _ := json.MarshalIndent(finalEntry, "", " ")
				ioutil.WriteFile(outputFileName, finalOutput, 0666)
			} else {
				finalOutput, _ := yaml.Marshal(finalEntry)
				ioutil.WriteFile(outputFileName, finalOutput, 0666)
			}
		}
	}
}
func GenerateCertKeyEntry(certPath, privKeyPath string, userID string, isJson bool, userEntryMap map[string]interface{}) {

	certBytes, _ := ioutil.ReadFile(certPath)
	keyBytes, _ := ioutil.ReadFile(privKeyPath)
	output := make(map[string]interface{})
	pemCert := make(map[string]string)
	if isJson {
		pemCert["pem"] = string(certBytes)
	} else {
		pemCert["path"] = certPath
	}
	output["cert"] = pemCert
	pemKey := make(map[string]interface{})
	if isJson {
		pemKey["pem"] = string(keyBytes)
	} else {
		pemKey["path"] = privKeyPath
	}
	output["key"] = pemKey
	userIDOnly := strings.Split(userID, "@")[0]
	userEntryMap[userIDOnly] = output

}
