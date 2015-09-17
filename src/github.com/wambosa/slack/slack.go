package slack

import (
	"fmt"
	"github.com/wambosa/net"
)

const Version = 1

const baseUrl = "https://slack.com/api/"

type SlackConfig struct {
	Token string
	Channel string
	LastRunTime string
}

var methodLinks map[string]string

func Init(conf SlackConfig) {

	methodLinks = map[string]string {
		"GetChannels": fmt.Sprintf("channels.list?token=%s", conf.Token),
		"GetAllChannelMessages": fmt.Sprintf("channels.history?token=%s&channel=", conf.Token),
		"GetAllDefaultChannelMessages": fmt.Sprintf("channels.history?token=%s&channel=%s", conf.Token, conf.Channel),
		"GetDefaultChannelMessagesSince": fmt.Sprintf("channels.history?token=%s&channel=%s&oldest=", conf.Token, conf.Channel),
		"GetDefaultChannelMessagesSinceLastRun": fmt.Sprintf("channels.history?token=%s&channel=%s&oldest=%s", conf.Token, conf.Channel, conf.LastRunTime),
		"PostMessageToDefaultChannel": fmt.Sprintf("chat.postMessage?token=%s&username=ChOPS&channel=%s&text=", conf.Token, conf.Channel),
	}
}

func ConvertSlackConfigToMap(conf SlackConfig) (map[string]interface{}){
	return map[string]interface{}{
		"token": conf.Token,
		"channel": conf.Channel,
		"lastRunTime":conf.LastRunTime,
	}
}

func GetChannels() (map[string]interface{}, error) {

	//todo: change the return type to either a slice of maps (or a map of maps)
	return simhttp.GetResponseAsMap(baseUrl + methodLinks["GetChannels"])
}

func GetAllChannelMessages(slackChannel string)(map[string]interface{}, error) {

	return simhttp.GetResponseAsMap(baseUrl + methodLinks["GetAllChannelMessages"] + slackChannel)
}

func GetAllDefaultChannelMessages()(map[string]interface{}, error) {

	return simhttp.GetResponseAsMap(baseUrl + methodLinks["GetAllDefaultChannelMessages"])
}

func GetDefaultChannelMessagesSince(unixStamp string)(map[string]interface{}, error) {

	return simhttp.GetResponseAsMap(baseUrl + methodLinks["GetDefaultChannelMessagesSince"] + unixStamp)
}

func GetDefaultChannelMessagesSinceLastRun()([]map[string]interface{}, error) {//todo: test out map[string]string

	response, err := simhttp.GetResponseAsMap(baseUrl + methodLinks["GetDefaultChannelMessagesSinceLastRun"])

	if(err != nil || response == nil){return nil, err}

	messages := make([]map[string]interface{}, len(response["messages"].([]interface{})))

	for i, message := range response["messages"].([]interface{}) {
		messages[i] = message.(map[string]interface{})
	}

	return messages, nil
}

func PostMessageToDefaultChannel(message string)(map[string]interface{}, error) {
	//todo: do some testing to determine string escaping needs.
	return simhttp.GetResponseAsMap(baseUrl + methodLinks["PostMessageToDefaultChannel"] + message)
}

//todo: Get User info using userId (this wil be done so that we can have a unified userbase across all apps)