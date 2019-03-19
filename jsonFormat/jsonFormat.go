package main

import (
	"encoding/json"
)

func main() {
	messageText := "This is my message"
	type ApsMessage struct {
		Alert string `json:"alert"`
		Sound string `json:"sound"`
	}
	type ApnsMessage struct {
		ApsMessage `json:"aps"`
	}
	apnsMessage := ApnsMessage{
		ApsMessage : ApsMessage{
			Alert:messageText,
			Sound:"default",
		},
	}
	apnsMessageJson, _ := json.Marshal(apnsMessage)
	apnsMessageString := string(apnsMessageJson)
	//var apnsMessageString = "junk"
	println (apnsMessageString)
	type MessageBody struct {
		Default string `json:"default"`
		APNS string `json:"APNS"`
		APNS_SANDBOX string `json:"APNS_SANDBOX"`
	}
	type Message struct {
		MessageBody
	}
	message := Message{
		MessageBody: MessageBody{
			Default:      messageText,
			APNS:         apnsMessageString,
			APNS_SANDBOX: apnsMessageString,
		},
	}
	messageJson, _ := json.Marshal(message)
	messageString := string(messageJson)
	println (messageString)
}
