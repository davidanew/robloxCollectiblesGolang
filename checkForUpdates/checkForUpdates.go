package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sns"
	"io/ioutil"
	"net/http"
	"time"
)

//set http client here so we can get a timeout
var myClient = &http.Client{Timeout: 10 * time.Second}

//const snsTopicArn = "arn:aws:sns:eu-west-1:168606352827:robloxCollectiblesTopic"

//Used for unmarshalling Json
type JsonType struct {
	Array []struct {
		AssetId int64
		Name    string
	}
}

//Used for writing and reading db
type Item struct {
	AssetId int64
}

//Data sent from cloudwatch
type Config struct {
	DbTableName string `json:"DbTableName"`
	SnsTopicArn string `json:"SnsTopicArn"`
}

func main() {
	lambda.Start(HandleRequest)
}

//request handler for cloudwatch
func HandleRequest(config Config) error {
	dbTableName := config.DbTableName
	snsTopicArn := config.SnsTopicArn
	fmt.Printf("Running checkForUpdates with %s and %s\n", dbTableName, snsTopicArn)
	err := checkForUpdates(dbTableName, snsTopicArn)
	if err != nil {
		fmt.Println("Error : %S", err.Error())
	}
	return err
}

// checks for new collectible item. If found then send a notification and updates the
// dynamo db
func checkForUpdates(dbTableName string, snsTopicArn string) error {
	const url = "https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=30"
	// Create aws session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	if err != nil {
		return fmt.Errorf("error creating session: %s", err.Error())
	}
	// create object for dynamo db and sns
	svcDb := dynamodb.New(sess)
	svcSns := sns.New(sess)
	//call getJson procedure to get one page of updates
	//thus is enough to see if anything is added
	arr, err := getJson(url, 1)
	if err != nil {
		return fmt.Errorf("error in getting JSON: %s", err.Error())
	}
	//the unmarshaled data has been put in 'arr'
	//top level of arr is an array. step through every item of this
	for _, item := range arr.Array {
		// 'item' will be a struct with the two elements below
		assetId := item.AssetId
		name := item.Name
		//check to see if the item is in the database
		found, err := checkDbForItem(assetId, svcDb, dbTableName)
		if err != nil {
			return fmt.Errorf("error in checking for item: %s", err.Error())
		}
		//found indicates wether the item has been found in the database
		fmt.Printf("Found is %t \n", found)
		//if it is not in the database then send notification and then add the new item
		if !found {
			//construct message to be sent using sns
			message := fmt.Sprintf("New item:%s", name)
			fmt.Printf("Message is %s", message)
			//send message
			err := publish(message, svcSns, snsTopicArn)
			if err != nil {
				return fmt.Errorf("error in sending notification: %s", err.Error())
			}
			//add item to database
			err = writeToDb(assetId, svcDb, dbTableName)
			if err != nil {
				return fmt.Errorf("error in writing to dB: %s", err.Error())
			}
		}
	}
	return nil
}

func getJson(urlBase string, pageNumber int) (JsonType, error) {
	//add page number to URL to select which page we want to get
	var url = fmt.Sprintf("%s&PageNumber=%d", urlBase, pageNumber)
	println("url is %s", url)
	//do the http request
	response, err := myClient.Get(url)
	if err != nil {
		//return empty object and error message
		return JsonType{}, fmt.Errorf("Error in http request: %s", err.Error())
	}
	//wait for http request to finish
	defer response.Body.Close()
	//read response body to get response data
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return JsonType{}, fmt.Errorf("Error in ioutil: %s", err.Error())
	}
	//unmarshal response data to JsonType structure 'arr'
	dataJson := responseData
	arr := JsonType{}
	err = json.Unmarshal([]byte(dataJson), &arr.Array)
	if err != nil {
		return JsonType{}, fmt.Errorf("Error in unmarshalling: %s", err.Error())
	}
	// if we have got this far then we have valid data in 'arr'
	return arr, nil
}

//check to see if the item is in the database
func checkDbForItem(assetId int64, svc *dynamodb.DynamoDB, dbTableName string) (bool, error) {
	//GetItem needs a string input
	assetIdString := fmt.Sprintf("%d", assetId)
	fmt.Printf("Looking for: %s , ", assetIdString)
	//run GetItem
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		//Topic name hardcoded at the moment
		TableName: aws.String(dbTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"AssetId": {
				N: aws.String(assetIdString),
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("error in GetItem: %s", err.Error())
	}
	//unmarshal the response into an Item structure
	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal Record: %s", err.Error())
	}
	//this is the unmarshalled return from GetItem
	fmt.Printf("Found : %v , ", item.AssetId)
	//if the item is not found the return will be zero
	if item.AssetId == 0 {
		fmt.Printf("Need to add %d ,", assetId)
		//asset id not found
		return false, nil
	}
	//asset id found
	return true, nil
}

//publish message to sns topic
func publish(message string, svc *sns.SNS, snsTopicArn string) error {
	params := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(snsTopicArn),
	}
	_, err := svc.Publish(params)
	return err
}

//this procedure writes a single asset id to the db
func writeToDb(assetId int64, svc *dynamodb.DynamoDB, dbTableName string) error {
	fmt.Printf("Processing %d \n", assetId)
	//the asset id needs to be put in an 'Item' struct so it can be marshalled into the db structure
	var item = new(Item)
	item.AssetId = assetId
	//marshall the item into the data structure for the db write
	//av will be the attribute value for PutItemInput
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("Got error marshalling map %s", err.Error())
	}
	//create PutItemInput
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(dbTableName),
	}
	//put the data in the db
	_, err = svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("Got error calling PutItem: %s", err.Error())
	}
	//no errors if we have reached this point
	return nil
}
