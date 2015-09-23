package main

import(
	"fmt"
	"github.com/wambosa/easydb"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"github.com/wambosa/confman"
	"github.com/wambosa/slack"
)

func main(){

	db, err := sql.Open("sqlite3", "C:\\work\\git\\project_juggernaut\\bin\\someday.db3")

	if(err != nil){fmt.Println("FATAL:", err)}

	defer db.Close()

	records, err := easydb.RunQuery(
		"sqlite3",
		"C:\\work\\git\\project_juggernaut\\bin\\someday.db3",
		"SELECT * FROM ReceivedMessage")

	fmt.Println(records, err)

	fmt.Println("\n................................................\n")

	mySlice := []interface{} {
		"foo",
		"man",
		"shu",
	}
	fmt.Println(mySlice...)

	fmt.Println("\n................................................\n")

	epoch := "1442602904.000000"[:10]

	var converted uint64

	converted, err = strconv.ParseUint(epoch, 10, 64)

	fmt.Println("converted to", converted, converted > 9000, err)

	TestConfLoad()
}



func TestConfLoad(){
	fmt.Println("\n................................................\n")


	confMap, err := confman.LoadJson("C:\\work\\git\\project_juggernaut\\bin\\slack.conf")

	fmt.Println("Conf Map", confMap, err)

	chans := make([]string, len(confMap["channels"].([]interface{})))

	for i, cha := range confMap["channels"].([]interface{}){
		chans[i] = cha.(string)}

	slackConf := slack.SlackConfig {
		Token: confMap["token"].(string),
		Channels: chans,
		LastRunTime: confMap["lastRunTime"].(string),
	}


	fmt.Println("chan length", len(slackConf.Channels))

	channels := slackConf.Channels

	raw, err := slack.GetChannelIds(slackConf.Token)

	fmt.Println("raw chans", raw, err)

	if len(channels) == 0 {
		channels, err = slack.GetChannelIds(slackConf.Token)}

	fmt.Println("slackConf", slackConf)

	fmt.Println("channels", channels, err)
}

