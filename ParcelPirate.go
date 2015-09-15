package main

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
	"../pkg/conf"
	"../pkg/slack"
	"../pkg/jugger"
	"fmt"
	"os"
	"encoding/json"
)

func main() {

	//todo: find a way to use relative paths
	pirateFile := "C:/work/git/notes/go/src/ParcelPirate.conf"
	slackFile := "C:/work/git/notes/go/src/Slack.conf"
	//todo: will need to load a different conf foreach message source (slack, email, lync, flowdock)
	pirateConf, err := conf.LoadJson(pirateFile)

	if err != nil {fatal(err)}

	connString := pirateConf["targetDatabase"].(string)

	_slackConf, err := conf.LoadJson(slackFile)

	if err != nil {fatal(err)}

	//todo: iterate over the conf property keys and apply them to the existing properties in custom struct
	slackConf := slack.SlackConfig {
		Token: _slackConf["token"].(string),
		Channel: _slackConf["channel"].(string),
		LastRunTime: _slackConf["lastRunTime"].(string),
	}

	slack.Init(slackConf) //ask sam about this

	messages, err := slack.GetDefaultChannelMessagesSinceLastRun()

	if err != nil {fatal(err)}

	slackConf.LastRunTime, err = CreateRecievedMessages(&messages)//PICK UP HERE



	if err != nil {fatal(err)}

	conf.SaveJson(slackFile, slackConf)

	//this is just debug reporting. remove later
	buf, _ := json.MarshalIndent(messages, "", "  ")
	fmt.Print(string(buf))
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func CreateReceivedMessages(messages *map[string]interface{})(jugger.ReceivedMessage, error) {

	latestTimeStamp := "9999999999.00000"

	if(len(messages) > 0) {

		latestTimeStamp = messages[0]["ts"].(string)
	}

	for _, message := range messages {

		// foreach message, create a message record in the sql database
		// PICK UP HERE look up sqlite and golang
	}

	//go ahead and convert the float back to string. because we only need a string from this point on
	return string(latestTimeStamp), nil
}