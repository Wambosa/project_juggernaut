package main

import (
	"os"
	"fmt"
	"time"
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
	user string
	text string
	ts string
}

type ProcessingMethod func()([]RawApiMessage, error)

func main() {

	ParcelProcessors := map[string]ProcessingMethod {
		"slack": ProcessSlackParcels,
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

		messages, err := ParcelProcessors[parcelTypeName]()

		if err != nil {fatal(fmt.Sprintf("ParcelProcessor %s failed", parcelTypeName), err)}

		if(len(messages) > 0) {

			receivedMessages, err := CreateReceivedMessages(&messages, parcelTypeId)

			if err != nil {fatal("Unable to convert []RawApiMessages to []jugger.ReceivedMessage", err)}

			err = SaveNewlyReceivedMessages(&receivedMessages)

			if err != nil {fatal("Failed to write to database", err)}

		}else{

			fmt.Printf("No messages found while crawling %s", parcelTypeName)
		}
	}
}

func ProcessSlackParcels()([]RawApiMessage, error){

	//todo: defautlt to db lookup, then failback to config file. if neither exist. exit.
	// "SELECT key, value FROM ConfigTable WHERE key IN('token', 'channel', 'lastRunTime') AND OwnerName = '%s'"
	confMap, err := confman.LoadJson(confman.GetThisFolder() + "\\slack.conf")

	if err != nil {return nil, err}

	slackConf := slack.SlackConfig {
		Token: confMap["token"].(string),
		Channel: confMap["channel"].(string),
		LastRunTime: confMap["lastRunTime"].(string),
	}

	slack.Init(slackConf) //todo: ask sam about this technique (proper to init a library?)

	//this return type needs to have generic keys across all types. use slack as the model citizen (design a struct or interface that deals with this)
	rawMessages, err := slack.GetDefaultChannelMessagesSinceLastRun()

	if err != nil {return nil, err}

	messages := make([]RawApiMessage, len(rawMessages))

	for i, mess := range rawMessages {

		if user, ok := mess["user"]; ok { //ensure that only real user messages are stored.
			messages[i] = RawApiMessage{
				user: user.(string),
				text: mess["text"].(string),
				ts:mess["ts"].(string),
			}
		}
	}

	if(len(messages) > 0){
		slackConf.LastRunTime = messages[0].ts 	//note: slack api returns message in newest/latest first order. time.Now().UTC().Format(time.RFC3339)
	}else{
		slackConf.LastRunTime = fmt.Sprintf("%v.000000", time.Now().UTC().Unix())
	}

	//todo: maybe error check this guy...
	confman.SaveJson(confman.GetThisFolder() + "\\slack.conf", slack.ConvertSlackConfigToMap(slackConf))

	return messages, nil
}

func fatal(myDescription string, err error) {
	fmt.Println("FATAL: ", myDescription)
	fmt.Println(err)
	os.Exit(1)
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

		slackInfo, err := slack.GetUserInfo(slackUser)

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
			userId = userIds[0]["UserId"].(int)

		}else{
			//then we do not have a relationship with juggernaut, create a userrecord and add in some preferences
			//todo: dig deeper by trying a first name, last name match ?

			res, err := easydb.Exec(db,
				`INSERT INTO User (NickName, FirstName, LastName) VALUES (?,?,?)`,
				slackInfo["name"].(string), profile["first_name"].(string), profile["last_name"].(string))

			if err != nil {return 0, err}

			id64, _ := res.LastInsertId()
			userId = int(id64)
			easydb.Exec(db,
				`INSERT INTO UserPreference (UserId, Key, Value) VALUES (?,?,?)`,
				fmt.Sprintf("%v",userId), "Email", profile["email"].(string))
		}

		//whether or not the user already existed, their slack data did not exist, so lets be sure to store it for next time.
		easydb.Exec(db,
			`INSERT INTO UserPreference (UserId, Key, Value) VALUES (?,?,?)`,
			userId, "SlackUser", slackUser)

	}else{
		userId = userIds[0]["UserId"].(int)
	}

	os.Exit(1)
	return userId, nil
}

func SaveNewlyReceivedMessages(messages *[]jugger.ReceivedMessage)(error){

	query := `
	INSERT INTO ReceivedMessage
	(ParcelTypeId, MessageText, UserId)
	Values(?, ?, ?)
	`
	db, err := sql.Open("sqlite3", ConnectionString)

	if(err != nil){return err}

	defer db.Close() //todo: understand this line better

	transaction, err := db.Begin()

	if(err != nil){return err}

	statement, err := transaction.Prepare(query)

	if(err != nil){return err}

	defer statement.Close()

	for _, message := range *messages {

		_, err := statement.Exec(
			message.ParcelTypeId,
			message.MessageText,
			message.UserId)

		if(err != nil){return err}
	}

	transaction.Commit()

	return nil
}