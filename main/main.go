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

var myClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	const url = "https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=30"
	const maxPageNumber = 100

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	svc := dynamodb.New(sess)

	if err != nil {
		fmt.Println("Error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	arr, err := getJson(url , 1)
	if err != nil {
		fmt.Println("In getting JSON:")
		fmt.Println(err.Error())
	}
	writeToDb(arr, svc)

    for pageNumber := 2; pageNumber <= maxPageNumber; pageNumber++ {
	    arr, err := getJson(url , pageNumber)
	    if err == nil {
			writeToDb(arr, svc)
		} else {
			fmt.Println(err.Error())
			continue
		}
    }
}

func writeToDb (arr JsonType , svc *dynamodb.DynamoDB ) {
	for _, item := range arr.Array {
		fmt.Printf("Processing %s \n", item.Name)
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("RobloxCollectibles"),
		}
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func getJson(urlBase string, pageNumber  int) (JsonType , error) {
	var url  = fmt.Sprintf("%s&PageNumber=%d", urlBase, pageNumber)
	println("url is %s" , url)
	response, err := myClient.Get(url)
	if err != nil {
		return JsonType{}, fmt.Errorf("Error in http request: %s", err.Error())
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return JsonType{}, fmt.Errorf("Error in ioutil: %s", err.Error())
	}
	dataJson := responseData
	arr := JsonType{}
	err = json.Unmarshal([]byte(dataJson), &arr.Array)
	if err != nil {
		return JsonType{}, fmt.Errorf("Error in unmarshalling: %s", err.Error())
	}
	return arr , nil
}

type JsonType struct {
	Array []struct{
		AssetId		   int64
		Name           string
	}
}

