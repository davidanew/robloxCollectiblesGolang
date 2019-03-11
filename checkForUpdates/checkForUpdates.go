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
//	const maxPageNumber = 100

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

	for _, item := range arr.Array {
		assetId := item.AssetId
//		println(assetId)
		checkDbForItem(assetId,svc)
	}
}


func checkDbForItem (assetId int64 , svc *dynamodb.DynamoDB) bool {

	assetIdString :=  fmt.Sprintf("%d",assetId);

	fmt.Printf("Looking for: %s , " , assetIdString)


	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("RobloxCollectibles"),
		Key: map[string]*dynamodb.AttributeValue{
			"AssetId": {
				N: aws.String(assetIdString),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		println("there is an error")
	}

//	println(result.Item)

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

//	println(item.AssetId)

	fmt.Printf("Found : %v \n" , item.AssetId)

	if item.AssetId == 0 {
		fmt.Printf("Need to add %d \n" , assetId)
	}


	/*
	if item == "" {
		fmt.Println("Could not find 'The Big New Movie' (2015)")
		return
	}
*/




	return false

	/*
	result, err := svc.GetItem(&dynamodb.GetItemInput{
    TableName: aws.String("Movies"),
    Key: map[string]*dynamodb.AttributeValue{
        "year": {
            N: aws.String("2015"),
        },
        "title": {
            S: aws.String("The Big New Movie"),
        },
    },
})

if err != nil {
    fmt.Println(err.Error())
    return
}

item := Item{}

err = dynamodbattribute.UnmarshalMap(result.Item, &item)

if err != nil {
    panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
}

if item.Title == "" {
    fmt.Println("Could not find 'The Big New Movie' (2015)")
    return
}

fmt.Println("Found item:")
fmt.Println("Year:  ", item.Year)
fmt.Println("Title: ", item.Title)
fmt.Println("Plot:  ", item.Info.Plot)
fmt.Println("Rating:", item.Info.Rating)


	*/

}


/*

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
			TableName: aws.String("RobloxCollectiblesTest"),
		}
		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

*/

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
//		Name           string
	}
}

type Item struct {
	AssetId		   int64
}