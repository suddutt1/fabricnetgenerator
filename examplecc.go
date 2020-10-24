package main

import (
	"io/ioutil"
)

//GenerateExampleCC Generate exmaple chain code
func GenerateExampleCC(version, basePath string) {
	ioutil.WriteFile(basePath+"kvstore.go", []byte(_base_cc_v1_template), 0666)
}

//GenerateExampleCCV2 generates a sample chaincode for v2.2.x
func GenerateExampleCCV2(version, basePath string) {
	ioutil.WriteFile(basePath+"kvstore.go", []byte(_version2x_cc_template), 0666)
	ioutil.WriteFile(basePath+"go.mod", []byte(_version2x_cc_gomod_template), 0666)
}

const _version2x_cc_gomod_template = `
module github.com/suddutt1/chaincode/simplesc

go 1.14

require (
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200728190242-9b3ae92d8664
	github.com/hyperledger/fabric-protos-go v0.0.0-20200923192742-3897341ac036
	github.com/sirupsen/logrus v1.7.0
)

`
const _base_cc_v1_template = `
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _MAIN_LOGGER = shim.NewLogger("BaseSmartContractMain")

type SmartContract struct {
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_MAIN_LOGGER.Infof("Inside the init method ")

	return shim.Success(nil)
}
func (sc *SmartContract) probe(stub shim.ChaincodeStubInterface) pb.Response {
	ts := ""
	_MAIN_LOGGER.Info("Inside probe method")
	tst, err := stub.GetTxTimestamp()
	if err == nil {
		ts = tst.String()
	}
	output := "{\"status\":\"Success\",\"ts\" : \"" + ts + "\" }"
	_MAIN_LOGGER.Info("Retuning " + output)
	return shim.Success([]byte(output))
}

func (sc *SmartContract) save(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 2 {
		return shim.Error("Invalid number of arguments")
	}
	key := args[0]
	value := args[1]
	txID := stub.GetTxID()
	dataToStore := map[string]string{
		"value":  value,
		"trxnId": txID,
		"id":     key,
	}
	jsonBytesToStore, _ := json.Marshal(dataToStore)
	stub.PutState(key, jsonBytesToStore)

	return shim.Success([]byte(jsonBytesToStore))
}
func (sc *SmartContract) saveKV(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	inputJSON := args[0]
	kvList := make([]map[string]string, 0)
	err := json.Unmarshal([]byte(inputJSON), &kvList)
	if err != nil {
		return shim.Error("Can not convert input JSON to valid input")
	}
	if len(kvList) == 0 {
		return shim.Error("Empty data provided")
	}
	for _, kv := range kvList {
		key := kv["key"]
		value := kv["value"]
		txID := stub.GetTxID()
		dataToStore := map[string]string{
			"value":  value,
			"trxnId": txID,
			"id":     key,
		}
		jsonBytesToStore, _ := json.Marshal(dataToStore)
		stub.PutState(key, jsonBytesToStore)
	}

	return shim.Success([]byte(fmt.Sprintf("%d records saved", len(kvList))))
}
func (sc *SmartContract) query(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	key := args[0]
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Success(nil)

	}

	return shim.Success(data)
}

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	case "probe":
		response = sc.probe(stub)
	case "save":
		response = sc.save(stub)
	case "saveKV":
		response = sc.saveKV(stub)
	case "query":
		response = sc.query(stub)
	default:
		response = shim.Error("Invalid action provoided")
	}
	return response
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_MAIN_LOGGER.Criticalf("Error starting  chaincode: %v", err)
	}
}


`

const _version2x_cc_template = `
package main

import (
	"encoding/json"
	"fmt"

	cid "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

var _mainLogger = logger.New()

//SmartContract is the simple smart contract status
type SmartContract struct {
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")

	return shim.Success(nil)
}
func (sc *SmartContract) probe(stub shim.ChaincodeStubInterface) pb.Response {
	ts := ""
	_mainLogger.Info("Inside probe method")
	tst, err := stub.GetTxTimestamp()
	if err == nil {
		ts = tst.String()
	}
	output := "{\"status\":\"Success\",\"ts\" : \"" + ts + "\" }"
	_mainLogger.Info("Retuning " + output)
	return shim.Success([]byte(output))
}

func (sc *SmartContract) save(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 2 {
		return shim.Error("Invalid number of arguments")
	}
	key := args[0]
	value := args[1]
	txID := stub.GetTxID()
	who, _ := cid.GetID(stub)
	clientCert, _ := cid.GetX509Certificate(stub)

	dataToStore := map[string]string{
		"value":  value,
		"trxnId": txID,
		"id":     key,
		"who":    who,
		"cn":     clientCert.Subject.CommonName,
	}
	jsonBytesToStore, _ := json.Marshal(dataToStore)
	stub.PutState(key, jsonBytesToStore)

	return shim.Success([]byte(jsonBytesToStore))
}
func (sc *SmartContract) saveKV(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	inputJSON := args[0]
	kvList := make([]map[string]string, 0)
	err := json.Unmarshal([]byte(inputJSON), &kvList)
	if err != nil {
		return shim.Error("Can not convert input JSON to valid input")
	}
	if len(kvList) == 0 {
		return shim.Error("Empty data provided")
	}
	for _, kv := range kvList {
		key := kv["key"]
		value := kv["value"]
		txID := stub.GetTxID()
		dataToStore := map[string]string{
			"value":  value,
			"trxnId": txID,
			"id":     key,
		}
		jsonBytesToStore, _ := json.Marshal(dataToStore)
		stub.PutState(key, jsonBytesToStore)
	}

	return shim.Success([]byte(fmt.Sprintf("%d records saved", len(kvList))))
}
func (sc *SmartContract) query(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	key := args[0]
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Success(nil)

	}

	return shim.Success(data)
}

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	case "probe":
		response = sc.probe(stub)
	case "save":
		response = sc.save(stub)
	case "saveKV":
		response = sc.saveKV(stub)
	case "query":
		response = sc.query(stub)
	default:
		response = shim.Error("Invalid action provoided")
	}
	return response
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_mainLogger.Errorf("Error starting  chaincode: %v", err)
	}
}


`
