package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)


var myClient = &http.Client{Timeout: 10 * time.Second}

func main() {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)

	if err != nil {
		fmt.Println("Error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	url := "https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=1"

	response, err := myClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}


	dataJson := responseData
	arr := JsonType{}
	_ = json.Unmarshal([]byte(dataJson), &arr.Array)
	//TODO: put in error checking
	fmt.Printf("Name is %d \n",arr.Array[0].Name)
	fmt.Printf("AssetId is %s \n",arr.Array[0].AssetId)
	fmt.Printf("Updated is %s \n",arr.Array[0].Updated)

	// Create DynamoDB client
	svc := dynamodb.New(sess)

//	assetId := arr.Array[0].AssetId
//	name := arr.Array[0].Name




//	items := getItems()

	// Add each item to Movies table:
//	for _, item := range items {
		av, err := dynamodbattribute.MarshalMap(arr.Array[0])

		if err != nil {
			fmt.Println("Got error marshalling map:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Create item in table Movies
		input := &dynamodb.PutItemInput{
			Item: av,
			TableName: aws.String("RobloxCollectibles"),
		}

		_, err = svc.PutItem(input)

		if err != nil {
			fmt.Println("Got error calling PutItem:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		//fmt.Println("Successfully added '",item.Title,"' (",item.Year,") to Movies table")
//	}
}

type JsonType struct {
	Array []struct{
		AssetId		   int64
		Name           string
		Updated        string
	}
}








