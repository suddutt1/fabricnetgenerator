package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GenerateReadme(config []byte, path string) {

	var buffer bytes.Buffer
	buffer.WriteString("Unzip and place the contents generated in a directory. \n")
	buffer.WriteString("The directory is referred as <network>.  \n")
	buffer.WriteString("All the commands must be executed from  <network> directory.  \n")

	buffer.WriteString("One time setup. The following commands \n")

	buffer.WriteString(" . ./downloadbin.sh \n")
	buffer.WriteString("\n To generated the cryto config and other configurations run " +
		"the following commands \n")
	buffer.WriteString(" . ./generateartifacts.sh \n\n")
	buffer.WriteString("To start the netowrk  \n\n")
	buffer.WriteString("  docker-compose up -d \n")
	buffer.WriteString("To build and join channel. Make sure that network is running \n\n")
	buffer.WriteString("   docker exec -it cli bash -e ./buildandjoinchannel.sh \n\n")
	buffer.WriteString("Install and instantiate chain codes\n")
	buffer.WriteString("Create the chain code directiory.\n")
	buffer.WriteString("  cd <network> \n")

	configMap := make(map[string]interface{})
	//chain code
	json.Unmarshal(config, &configMap)
	ccList := configMap["chaincodes"].([]interface{})
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("  mkdir -p chaincode/%s \n", ccDetailsMap["src"]))
	}
	buffer.WriteString("Copy the chain code files in the respectivive directories \n")

	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})

		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_install.sh\n", ccDetailsMap["ccid"].(string)))
	}
	buffer.WriteString("To update the chain code , first update the chain code source files in the above mentioned directory.\n" + "Then run the following commands as appropriate\n\n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_update.sh <version>\n", ccDetailsMap["ccid"].(string)))
	}
	ioutil.WriteFile(path+"/README.txt", buffer.Bytes(), 0666)

}
