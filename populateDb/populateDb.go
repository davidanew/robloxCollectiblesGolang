package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

//set http client here so we can get a timeout
var myClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	const url = "https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=30"
	//currently there are 1986 collectibles according to the roblox web site
	//this is 67 pages
	const maxPageNumber = 100
	// Create aws session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	// create object for dynamo db and sns
	// TODO: this should be after error checking
	svc := dynamodb.New(sess)
	if err != nil {
		fmt.Println("Error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//call getJson procedure to get one page of updates
	arr, err := getJson(url , 1)
	if err != nil {
		//TODO: this should be a fatal error
		fmt.Println("In getting JSON:")
		fmt.Println(err.Error())
	}
	//write this data to db
	writeToDb(arr, svc)

	//now go through the other pages
	//terminate loop if maxPageNumber has been reached
	//or getJson returns an error
    for pageNumber := 2; pageNumber <= maxPageNumber; pageNumber++ {
	    arr, err := getJson(url , pageNumber)
	    if err == nil {
			writeToDb(arr, svc)
		} else {
			//If there is an error here we assume it is the end of the data and stop reading data
			fmt.Println(err.Error())
			continue
		}
    }
}

//this function is passed the full unmarshalled json
//it writes all the asset ids from this json to the db
func writeToDb (arr JsonType , svc *dynamodb.DynamoDB ) {
	//top level of arr is an array. step through every item of this
	for _, item := range arr.Array {
		//each 'item' will be a sctruct with one element for AssetId
		fmt.Printf("Processing %d \n", item.AssetId)
		//marshall the item into the data structure for the db write
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		//package data for PutItem
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("RobloxCollectiblesTest"),
		}
		//put the data in the db
		_, err = svc.PutItem(input)
		if err != nil {
			//TODO: structure returns in the same way as checkForUpdates
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func getJson(urlBase string, pageNumber  int) (JsonType , error) {
	//add page number to URL to select which page we want to get
	var url  = fmt.Sprintf("%s&PageNumber=%d", urlBase, pageNumber)
	println("url is %s" , url)
	//do the http request
	response, err := myClient.Get(url)
	if err != nil {
		//return empty object and error message
		return JsonType{}, fmt.Errorf("Error in http request: %s", err.Error())
	}
	//wait for http request to finish
	defer response.Body.Close()
	//TODO: comment here
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
	return arr , nil
}

//Used for unmarshalling Json
type JsonType struct {
	Array []struct{
		AssetId		   int64
//		Name           string
	}
}

