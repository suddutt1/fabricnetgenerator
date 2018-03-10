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
	buffer.WriteString("========First time setup instructions ( START)============= \n")
	buffer.WriteString("First time setup. Run the following commands \n")
	buffer.WriteString(" 1. Download the binaries \n")
	buffer.WriteString(" . ./downloadbin.sh \n")
	buffer.WriteString("\n 2. To Generate the cryto config and other configurations\n ")
	buffer.WriteString(" . ./generateartifacts.sh \n\n")
	buffer.WriteString("\n 3. To start the netowrk  \n\n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose up -d \n")

	buffer.WriteString("\n 4. To build and join channel. Make sure that network is running \n\n")
	buffer.WriteString("   docker exec -it cli bash -e ./buildandjoinchannel.sh \n\n")
	buffer.WriteString("\n 5. Install and instantiate chain codes\n")
	buffer.WriteString("\n 5a. Create the chain code directiory.\n")
	buffer.WriteString("  cd <network> \n")

	configMap := make(map[string]interface{})
	//chain code
	json.Unmarshal(config, &configMap)
	ccList := configMap["chaincodes"].([]interface{})
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("  mkdir -p chaincode/%s \n", ccDetailsMap["src"]))
	}
	buffer.WriteString("\n 5.b Copy the chain code files in the respectivive directories \n")
	buffer.WriteString("\n 5.c Install and intantiate the chain codes \n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})

		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_install.sh\n", ccDetailsMap["ccid"].(string)))
	}
	buffer.WriteString("========First time setup instructions ( END)============= \n")
	buffer.WriteString("\n\n========When chain code is modified ( START)============= \n")
	buffer.WriteString(" To update the chain code , first update the chain code source files in the above mentioned directory.\n" + "Then run the following commands as appropriate\n\n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_update.sh <version>\n", ccDetailsMap["ccid"].(string)))
	}
	buffer.WriteString("========When chain code is modified ( END)============= \n")
	buffer.WriteString("\n\n========To bring up an existing network ( START)============= \n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose up -d \n")
	buffer.WriteString("========To bring up an existing network ( END)============= \n")
	buffer.WriteString("\n\n========To destory  an existing network ( START)============= \n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose down \n")
	buffer.WriteString(" If you are stoping a network using the above commands , " +
		"\n to start the network again , you have to execute steps 2,3,4 & 5a,5c of the first time setup.\n ")

	buffer.WriteString("\n\n========To destory  an existing network ( END)============= \n")
	ioutil.WriteFile(path+"/README.txt", buffer.Bytes(), 0666)

}
