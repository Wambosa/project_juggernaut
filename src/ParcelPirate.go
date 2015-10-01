package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"database/sql"
	"github.com/wambosa/easydb"
	"github.com/wambosa/slack"
	"github.com/wambosa/jugger"
	"github.com/wambosa/confman"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ConnectionString string
)

type RawApiMessage struct {
	channel string
	user string
	text string
	ts string
}

type ProcessingMethod func()([]RawApiMessage, error)

func main() {
	fmt.Println("Parcel Pirate 2")
	//todo : ascii art
	ParcelSnatcher := map[string]ProcessingMethod {
		"slack": SnatchSlackParcels,
	}

	pirateConf, err := confman.LoadJson(fmt.Sprint(confman.GetThisFolder(), "\\ParcelPirate.conf"))

	if err != nil {fatal("ParcelPirate Conf Failed to Load", err)}

	ConnectionString = pirateConf["targetDatabase"].(string)

	supportedParcelTypes := map[string]int{ // := pirateConf["parcelTypes"].([]string)
		"slack": 1,
		//"lync":2,
		//"smtp":3,
		//"sms":4,
		//"voice":5,
		//"flowdock":6,
	}

	for parcelTypeName, parcelTypeId := range supportedParcelTypes {

		parcels, err := ParcelSnatcher[parcelTypeName]()

		if err != nil {fatal(fmt.Sprintf("ParcelProcessor %s failed", parcelTypeName), err)}

		if(len(parcels) > 0) {

			fmt.Printf("snatched %v parcels from %s\n", len(parcels), parcelTypeName)

			receivedMessages, err := CreateReceivedMessages(&parcels, parcelTypeId)

			if err != nil {fatal("Unable to convert []RawApiMessages to []jugger.ReceivedMessage", err)}

			fmt.Printf("burying %s booty!\n", parcelTypeName)

			err = SaveNewReceivedMessages(&receivedMessages)

			if err != nil {fatal("Unable to save new []jugger.ReceivedMessage in database", err)}

		}else{

			fmt.Printf("No parcels found while sailing near %s\n", parcelTypeName)
		}
	}

	fmt.Println("done")
}

func SnatchSlackParcels()([]RawApiMessage, error){

	//todo: defautlt to db lookup, then failback to config file. if neither exist. exit.
	// "SELECT key, value FROM ConfigTable WHERE key IN('token', 'channel', 'lastRunTime') AND OwnerName = '%s'"
	confMap, err := confman.LoadJson(confman.GetThisFolder() + "\\slack.conf")

	if err != nil {return nil, err}

	chanSlice := make([]string, len(confMap["channels"].([]interface{})))

	for i, cha := range confMap["channels"].([]interface{}) {
		chanSlice[i] = cha.(string)}

	slackConf := slack.SlackConfig {
		Token: confMap["token"].(string),
		Channels: chanSlice,
		LastRunTime: confMap["lastRunTime"].(string),
	}

	channelIds := slackConf.Channels

	if (len(channelIds) == 0) {
		channelIds, err = slack.GetChannelIds(slackConf.Token)
		if err != nil {return nil, err}
	}

	allMessages := make([]RawApiMessage, 0)
	oldestTime := slackConf.LastRunTime

	for _, channelId := range channelIds {

		// this return type needs to have generic keys across all types.
		// use slack as the model citizen (design a struct or interface that deals with this)
		slackMessages, err := slack.GetChannelMessagesSinceLastRun(slackConf.Token, channelId, oldestTime)

		if err != nil {return nil, err}

		channelMessages := make([]RawApiMessage, 0)

		for _, mess := range slackMessages {
			if _, ok := mess["user"]; ok { //ensure that only real user messages are stored. (slack bot messages do not have a user property)
					channelMessages = append(channelMessages, RawApiMessage{
						channel: channelId, //this will be the current channel in this for loop
						user: mess["user"].(string),
						text: mess["text"].(string),
						ts:mess["ts"].(string),
					})
			}
		}

		if(len(channelMessages) > 0){

			allMessages = append(allMessages, channelMessages...)

			messageEpoch,_ := strconv.ParseUint(channelMessages[0].ts[:10], 10, 64)
			configEpoch,_ := strconv.ParseUint(slackConf.LastRunTime[:10], 10, 64)

			if( messageEpoch > configEpoch) {
				slackConf.LastRunTime = channelMessages[0].ts}
		}
	}

	if(len(allMessages) == 0) {
		slackConf.LastRunTime = fmt.Sprintf("%v.000000", time.Now().UTC().Unix())}

	//todo: maybe error check this guy...
	confman.SaveJson(confman.GetThisFolder() + "\\slack.conf", slack.ConvertSlackConfigToMap(slackConf))

	return allMessages, nil
}

func CreateReceivedMessages(messages *[]RawApiMessage, parcelType int)([]jugger.ReceivedMessage, error) {

	receivedMessages := make([]jugger.ReceivedMessage, len(*messages), len(*messages))

	for i, message := range *messages {

		userId, err := FindUserIdBySlackUser(message.user, parcelType)

		if(err != nil){return nil, err}

		receivedMessages[i] = jugger.ReceivedMessage{
			ReceivedMessageId: 0,
			ParcelTypeId: parcelType,
			MessageText: message.text,
			UserId: userId,
			JobId: 0,
			ParseStatusId: 1,
			CreatedOn: time.Now().UTC(),
			LastUpdated: time.Now().UTC(),
			Metadata: []jugger.ReceivedMessageMetadata{
				jugger.ReceivedMessageMetadata{
					Key:"Channel",
					Value: message.channel},
			},
		}
	}

	return receivedMessages, nil
}

func FindUserIdBySlackUser(slackUser string, parcelType int)(int, error){
	//note: this has the potential to get screwed up if my earching returns more than one result.
	// i am counting on the fact that people cannot share email addresses and that slack user ids never get recyled
	// if either of those assumptions is bad, then we can get screwed up references

	bestTryQuery := `
	SELECT UserId
	FROM UserPreference
	WHERE Key = 'SlackUser'
	AND Value LIKE '%s'
	`
	db, err := sql.Open("sqlite3", ConnectionString)

	if(err != nil){return 0, err}

	defer db.Close()

	userIds, err := easydb.Query(db, fmt.Sprintf(bestTryQuery, slackUser))

	if(err != nil){return 0, err}

	var userId int

	//then we either have not yet seen this user in slack, or they do not have relationship with juggernaut
	if(len(userIds) == 0){

		slackInfo, err := slack.GetUserInfoWithCachedToken(slackUser)

		if err != nil {return 0, err}

		profile := slackInfo["profile"].(map[string]interface{})

		findByEmailQuery := `
		SELECT UserId
		FROM UserPreference
		WHERE Key = 'Email'
		AND Value = '%s'
		`
		userIds, err = easydb.Query(db, fmt.Sprintf(findByEmailQuery, profile["email"].(string)))

		if err != nil {return 0, err}

		if(len(userIds) > 0){
			//we have a relationship, but did not recognize the user in this context at first.
			userId = int(userIds[0]["UserId"].(int64))

		}else{
			//then we do not have a relationship with juggernaut, create a userrecord and add in some preferences
			//todo: dig deeper by trying a first name, last name match ?

			var nickname, firstName, lastName string

			if val, ok := slackInfo["name"]; ok {nickname = val.(string)}
			if val, ok := profile["first_name"]; ok {firstName = val.(string)}
			if val, ok := profile["last_name"]; ok {lastName = val.(string)}

			res, err := easydb.Exec(db,
				`INSERT INTO User (NickName, FirstName, LastName) VALUES (?,?,?)`,
				nickname, firstName, lastName)

			if err != nil {return 0, err}

			lastId, _ := res.LastInsertId()
			userId = int(lastId)
			easydb.Exec(db,
				`INSERT INTO UserPreference (UserId, Key, Value) VALUES (?,?,?)`,
				fmt.Sprintf("%v",userId), "Email", profile["email"].(string))
		}

		//whether or not the user already existed, their slack data did not exist, so lets be sure to store it for next time.
		easydb.Exec(db,
			`INSERT INTO UserPreference (UserId, Key, Value) VALUES (?,?,?)`,
			userId, "SlackUser", slackUser)

	}else{
		userId = int(userIds[0]["UserId"].(int64))
	}

	return userId, nil
}

func SaveNewReceivedMessages(messages *[]jugger.ReceivedMessage)(error){

	messageQuery := `
	INSERT INTO ReceivedMessage
	(ParcelTypeId, MessageText, UserId)
	Values(?, ?, ?)
	`
	metaQuery := `
	INSERT INTO ReceivedMessageMetadata
	(ReceivedMessageId, Key, Value)
	Values(?, ?, ?)
	`

	db, err := sql.Open("sqlite3", ConnectionString)

	if(err != nil){return err}

	defer db.Close() //todo: understand this line better

	for _, message := range *messages {

		res, err := easydb.Exec(db, messageQuery, message.ParcelTypeId, message.MessageText, message.UserId)

		//todo: (improve performance and io by using a view to abstract the two tables into 1)

		if(err != nil){return err}

		lastId, _ := res.LastInsertId()

		for _, meta := range message.Metadata {

			_, err := easydb.Exec(db, metaQuery, lastId, meta.Key, meta.Value)

			if(err != nil){return err}
		}
	}

	return nil
}

func fatal(myDescription string, err error) {
	fmt.Println("FATAL: ", myDescription)
	fmt.Println(err)
	os.Exit(1)
}