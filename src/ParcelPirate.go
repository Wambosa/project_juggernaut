package main

import (
	"os"
	"fmt"
	"time"
	"database/sql"
	"github.com/wambosa/confman"
	"github.com/wambosa/slack"
	"github.com/wambosa/jugger"
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

type ProcessingMethod func(map[string]confman.LoadFunc, map[string]string)([]RawApiMessage, error)

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

		//todo: this is conf load section still needs a bit more versatility with what the key names are. likely different for each different source. store that somewhere...
		getConfFuncs := map[string]confman.LoadFunc {
			"db": confman.LoadSqlite,
			"json":confman.LoadJson,
		}
		getConfArgs := map[string]string {
			"db": fmt.Sprintf("SELECT key, value FROM ConfigTable WHERE key IN('token', 'channel', 'lastRunTime') AND OwnerName = '%s'", parcelTypeName),
			"json":fmt.Sprintf("%s\\%s.conf", confman.GetThisFolder(), parcelTypeName),
		}

		//messages, err := ParcelProcessors[parcelTypeName].(func(map[string]confman.LoadFunc, map[string]string))(getConfFuncs, getConfArgs)
		messages, err := ParcelProcessors[parcelTypeName](getConfFuncs, getConfArgs)

		if err != nil {fatal(fmt.Sprintf("ParcelProcessor %s failed", parcelTypeName), err)}

		if(len(messages) > 0){

			receivedMessages, err := CreateReceivedMessages(&messages, parcelTypeId)

			if err != nil {fatal("Unable to convert RawApiMessages to []jugger.ReceivedMessage", err)}

			SaveNewlyReceivedMessages(&receivedMessages)
		}
	}
}

func ProcessSlackParcels(getConfFuncs map[string]confman.LoadFunc, getConfArgs map[string]string)([]RawApiMessage, error){

	confMap, err := getConfFuncs["json"](getConfArgs["json"])

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
		slackConf.LastRunTime = string(time.Now().UTC().Unix())
	}

	//todo: maybe error check this guy.. also need to accomodate database saving. so may need to add this to the func map
	confman.SaveJson(getConfArgs["json"], slack.ConvertSlackConfigToMap(slackConf))

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

		userId, err := FindUserId(message.user, parcelType)

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

func FindUserId(magneticValue string, parcelType int)(int, error){

	//search smartly for the user. you will need to reach to the user from their user data, then compare the name, email or whatever we can find
	//depending on the parcel source, i may be able to query against the api again in order to discover what the source knows about the user, the more data i can get, the less i have to assume.
	//try to match using email first, then username, then first last name combo

	//we must return a userId no matter what.

	query := `
	SELECT UserId
	FROM UserPreference
	WHERE Value LIKE '%s'
	`
	db, err := sql.Open("sqlite3", ConnectionString)

	if(err != nil){return 0, err}

	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(query, magneticValue))

	if(err != nil){return 0, err}

	var userId int

	rowCount := 0

	for rows.Next() {
		//this is where we have to get smart.
		//if there are multiple rows, then something is likely wrong. wee need to pick the best option. for now ignore this possibility
		rows.Scan(&userId)

		if(rowCount > 0){

			fmt.Println("FinsUserId has found multiple matches for ", magneticValue, userId)
		}
	}

	return userId, nil
}

func SaveNewlyReceivedMessages(messages *[]jugger.ReceivedMessage)(error){

	query := `
	INSERT INTO ReceivedMessage
	(ParcelTypeId, MessageText, UserId)
	Values(?, ?, ?, ?)
	`
	db, err := sql.Open("sqlite3", ConnectionString)

	if(err != nil){return err}

	defer db.Close() //todo: understand this line better

	transaction, err := db.Begin()

	if(err != nil){return err}

	q, err := transaction.Prepare(query)

	if(err != nil){return err}

	defer q.Close()

	for _, message := range *messages {

		_, err = q.Exec(
			message.ParcelTypeId,
			message.MessageText,
			message.UserId)

		if(err != nil){return err}
	}

	transaction.Commit()

	return nil
}