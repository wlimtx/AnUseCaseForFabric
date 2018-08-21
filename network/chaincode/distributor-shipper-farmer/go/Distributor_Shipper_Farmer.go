/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble","marble1","blue","35","tom","99"]}'
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble","marble2","red","50","tom","102"]}'
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["initMarble","marble3","blue","70","tom","103"]}'
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C mychannel -n marblesp -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["readMarblePrivateDetails","marble1"]}'
// peer chaincode query -C mychannel -n marblesp -c '{"Args":["getMarblesByRange","marble1","marble3"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Or for indexes on private data
// collections, in a META-INF/statedb/couchdb/collections/<collection_name>/indexes directory.
// Each index must be defined in its own text file with extension *.json with the index
// definition formatted in JSON following the CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02_private example chaincode demonstrates a packaged index which you
// can find in META-INF/statedb/couchdb/collection/collectionMarbles/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory or
// META-INF/statedb/couchdb/collections/<collection_name>/indexes directory, for packaging
// and deployment to managed environments.
//
// In the examples below you can find index definitions that support marbles02_private
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":["data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}

// Index for docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C mychannel -n marblesp -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// DistributorFarmerShipperChainCode example simple Chaincode implementation
type DistributorFarmerShipperChainCode struct {
}

type fruitTransactionDetail struct {
	ObjectType	string	`json:"docType"` //docType is used to distinguish the various types of objects in state database
	Id        	int64 	`json:"id"`
	Name      	string	`json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Type      	string	`json:"type"`
	Price     	string	`json:"price"`
	Weight    	string	`json:"weight"`
	Level     	string	`json:"level"`
}

type deliverFeeDetail struct {
	ObjectType	string	`json:"docType"`
	Id        	int64 	`json:"id"`
	Price     	string	`json:"price"`
	Weight    	string	`json:"weight"`
	From      	string	`json:"from"`
	To        	string	`json:"to"`
	DeliverMan	string	`json:"deliverman"`
	SignMan   	string	`json:"signman"`
	SignTel   	string	`json:"signtel"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(DistributorFarmerShipperChainCode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *DistributorFarmerShipperChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *DistributorFarmerShipperChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	switch function {
	case "newFruitTx":
		//read a marble
		return t.newFruitTx(stub, args)
	case "newDeliverFeeTx":
		//read a marble private details
		return t.newDeliverFeeTx(stub, args)
	case "readFruitTx":
		//change owner of a specific marble
		return t.readFruitTx(stub, args)
	case "readDeliverFeeTx":
		//transfer all marbles of a certain color
		return t.readDeliverFeeTx(stub, args)
	case "richQueryFruitTx":
		//delete a marble
		return t.richQueryFruitTx(stub, args)
	case "richQueryDeliverFeeTx":
		//find marbles for owner X using rich query
		return t.richQueryDeliverFeeTx(stub, args)
	case "getFruitTxByRange":
		//find marbles based on an ad hoc rich query
		return t.getFruitTxByRange(stub, args)
	case "getDeliverFeeTxByRange":
		//find marbles based on an ad hoc rich query
		return t.getDeliverFeeTxByRange(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================
// newFruitTx - create a new Fruit Transaction, store into chain code state
// ============================================================
func (t *DistributorFarmerShipperChainCode) newFruitTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//	ID 	名字  类型  单价    重量 	  等级
	//	id	name type  price  weight  level
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init Fruit Transaction")
	if len(args[0]) == 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) == 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) == 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) == 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) == 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) == 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	fruitId, err := strconv.ParseInt(args[0], 10, 64)
	if err!= nil {
		return shim.Error("1th argument must be a numeric string")
	}
	fruitName := args[1]
	fruitType := args[2]
	fruitPrice := args[3]
	fruitWeight := args[4]
	fruitLevel := args[5]

	// ==== Check if fruitTransactionDetail already exists ====
	fruitTxAsBytes, err := stub.GetPrivateData("fruitTransactionDetails", args[0])
	if err != nil {
		return shim.Error("Failed to get fruitTransactionDetail: " + err.Error())
	} else if fruitTxAsBytes != nil {
		fmt.Println("This fruitTransactionDetail already exists: " + args[0])
		return shim.Error("This fruitTransactionDetail already exists: " + args[0])
	}

	// ==== Create fruitTransactionDetail object and marshal to JSON ====
	objectType := "fruitTransactionDetail"
	fruitTransactionDetail := &fruitTransactionDetail{objectType,fruitId, fruitName, fruitType,fruitPrice, fruitWeight, fruitLevel}
	fruitTransactionDetailAsBytes, err := json.Marshal(fruitTransactionDetail)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the fruitTransactionDetail json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + fruitName + `", "fruitType": "` + fruitType + `", "fruitWeight": ` + strconv.Itoa(fruitWeight) + `, "fruitPrice": "` + fruitPrice + `"}`
	//fruitTransactionDetailAsBytes := []byte(str)

	// === Save fruitTransactionDetail to state ===
	err = stub.PutPrivateData("fruitTransactionDetails", args[0], fruitTransactionDetailAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//// ==== Save fruitTransactionDetail private details ====
	//objectType = "marblePrivateDetails"
	//marblePrivateDetails := &marblePrivateDetails{objectType, fruitName, fruitLevel}
	//marblePrivateDetailsBytes, err := json.Marshal(marblePrivateDetails)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//err = stub.PutPrivateData("collectionMarblePrivateDetails", fruitName, marblePrivateDetailsBytes)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}

	////  ==== Index the fruitTransactionDetail to enable fruitType-based range queries, e.g. return all blue marbles ====
	////  An 'index' is a normal key/value entry in state.
	////  The key is a composite key, with the elements that you want to range query on listed first.
	////  In our case, the composite key is based on indexName~fruitType~name.
	////  This will enable very efficient state range queries based on composite keys matching indexName~fruitType~*
	//indexName := "fruitType~name"
	//colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{fruitTransactionDetail.Color, fruitTransactionDetail.Name})
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	////  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the fruitTransactionDetail.
	////  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	//value := []byte{0x00}
	//stub.PutPrivateData("collectionMarbles", colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init fruitTransactionDetail")
	return shim.Success(nil)
}





// ============================================================
// newDeliverFeeTx - create a new Deliver Fee, store into chain code state
// ============================================================
func (t *DistributorFarmerShipperChainCode) newDeliverFeeTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// Id	Price	Weight	From	To	DeliverMan	SignMan	SignTel


	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init Deliver Fee Transaction")
	if len(args[0]) == 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) == 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) == 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) == 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) == 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) == 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) == 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) == 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	feeId, err := strconv.ParseInt(args[0], 10, 64)
	if err!= nil {
		return shim.Error("1th argument must be a numeric string")
	}
	Price := args[1]
	Weight := args[2]
	From := args[3]
	To := args[4]
	DeliverMan := args[5]
	SignMan := args[6]
	SignTel := args[7]

	// ==== Check if deliverFeeDetail already exists ====
	feeTxAsBytes, err := stub.GetPrivateData("deliverFeeDetails", args[0])
	if err != nil {
		return shim.Error("Failed to get deliverFeeDetails: " + err.Error())
	} else if feeTxAsBytes != nil {
		fmt.Println("This deliverFeeDetails already exists: " + args[0])
		return shim.Error("This deliverFeeDetails already exists: " + args[0])
	}

	// ==== Create deliverFeeDetail object and marshal to JSON ====
	objectType := "deliverFeeDetail"
	deliverFeeDetail := &deliverFeeDetail{objectType, feeId, Price, Weight, From, To, DeliverMan,SignMan, SignTel}

	deliverFeeDetailAsBytes, err := json.Marshal(deliverFeeDetail)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save deliverFeeDetail to state ===
	err = stub.PutPrivateData("deliverFeeDetails", args[0], deliverFeeDetailAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init deliverFeeDetail")
	return shim.Success(nil)
}




// ===============================================
// readMarble - read a fruit tx from chaincode state
// ===============================================
func (t *DistributorFarmerShipperChainCode) readFruitTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var id, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting id of the tx to query")
	}

	id = args[0]
	valAsBytes, err := stub.GetPrivateData("fruitTransactionDetails", id) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error\":\"tx does not exist: " + id + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsBytes)
}

func (t *DistributorFarmerShipperChainCode) readDeliverFeeTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var id, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting id of the tx to query")
	}

	id = args[0]
	valAsBytes, err := stub.GetPrivateData("deliverFeeDetails", id) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error\":\"tx does not exist: " + id + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsBytes)
}



// ===========================================================================================
// getMarblesByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *DistributorFarmerShipperChainCode) getFruitTxByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]


	queryResults, err := getTxByRange(stub,"fruitTransactionDetails", startKey, endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *DistributorFarmerShipperChainCode) getDeliverFeeTxByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]


	queryResults, err := getTxByRange(stub,"deliverFeeDetails", startKey, endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getTxByRange(stub shim.ChaincodeStubInterface, collection string, startKey string, endKey string) ([]byte, error) {

	resultsIterator, err := stub.GetPrivateDataByRange(collection, startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getFruitTxByRange queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}



// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *DistributorFarmerShipperChainCode) richQueryFruitTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub,"fruitTransactionDetails", queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}
func (t *DistributorFarmerShipperChainCode) richQueryDeliverFeeTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub,"deliverFeeDetails", queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, PDCName string,queryString string) ([]byte, error) {


	fmt.Printf("- PDCName:%s queryString:\n%s\n",PDCName, queryString)

	resultsIterator, err := stub.GetPrivateDataQueryResult(PDCName, queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
