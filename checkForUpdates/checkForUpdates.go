package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sns"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}
const snsTopicArn = "arn:aws:sns:eu-west-1:168606352827:robloxCollectiblesTopic"

func main() {
	const url = "https://search.roblox.com/catalog/json?SortType=RecentlyUpdated&IncludeNotForSale=false&Category=Collectibles&ResultsPerPage=30"
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	if err != nil {
		fmt.Println("Error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	svcDb := dynamodb.New(sess)
	svcSns := sns.New(sess)
	arr, err := getJson(url , 1)
	if err != nil {
		fmt.Println("Error in getting JSON:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for _, item := range arr.Array {
		assetId := item.AssetId
		name := item.Name
		found , err := checkDbForItem(assetId, svcDb)
		if err != nil {
			fmt.Println("Error in checking for item:")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Found is %t \n" , found)
		if !found {
			message := fmt.Sprintf("New item:%s", name)
			fmt.Printf("Message is %s", message)
			err := publish(message , svcSns)
			if err != nil {
				fmt.Println("Error in sending notification:")
				fmt.Println(err.Error())
				os.Exit(1)
			}
			err = writeToDb(assetId, svcDb)
			if err != nil {
				fmt.Println("Error in writing to dB:")
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
}

func writeToDb (assetId int64 , svc *dynamodb.DynamoDB ) error {
	fmt.Printf("Processing %d \n", assetId)
	var item = new(Item)
	item.AssetId = assetId

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("Got error marshalling map %s", err.Error())
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("RobloxCollectibles"),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("Got error calling PutItem: %s", err.Error())
	}
	return nil
}

func publish (message string, svc *sns.SNS) (error) {
	params := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(snsTopicArn),
	}
	_, err := svc.Publish(params)
	return err
}

func checkDbForItem (assetId int64 , svc *dynamodb.DynamoDB) (bool ,error) {
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
		return false, fmt.Errorf("error in GetItem: %s", err.Error())

	}
	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal Record: %s", err.Error())
	}
	fmt.Printf("Found : %v , " , item.AssetId)
	if item.AssetId == 0 {
		fmt.Printf("Need to add %d ," , assetId)
		return false , nil
	}
	return true , nil
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
		Name		   string

	}
}

type Item struct {
	AssetId		   int64
}