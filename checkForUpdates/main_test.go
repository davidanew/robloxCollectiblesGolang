package main

import "testing"

func TestCheckForUpdates(t *testing.T) {
	println("running test")
	err := checkForUpdates("RobloxCollectiblesTest","arn:aws:sns:eu-west-1:168606352827:robloxCollectiblesTopic")
	if err != nil {
		t.Errorf("Test failed : %s " , err.Error())
	}
}

