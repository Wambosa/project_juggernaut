package slack

import (
	"fmt"
	"github.com/wambosa/net"
)

const Version = 1

const baseUrl = "https://slack.com/api/"

var cachedToken string

type SlackConfig struct {
	Token string
	Channels []string
	LastRunTime string
}

func ConvertSlackConfigToMap(conf SlackConfig) (map[string]interface{}){
	return map[string]interface{}{
		"token": conf.Token,
		"channels": conf.Channels, //todo: test this
		"lastRunTime":conf.LastRunTime,
	}
}

func GetChannels(token string) ([]map[string]interface{}, error) {

	cachedToken = token

	method := fmt.Sprintf("%schannels.list?token=%s", baseUrl, token)

	response, err := simhttp.GetResponseAsMap(method)

	if err != nil || response == nil {return nil, err}

	channels := make([]map[string]interface{}, len(response["channels"].([]interface{})))

	for i, channel := range response["channels"].([]interface{}){
		channels[i] = channel.(map[string]interface{})}

	return channels, nil
}

func GetChannelIds(token string) ([]string, error) {

	cachedToken = token

	chans, err := GetChannels(token)

	if err != nil {return nil, err}

	channelIds := make([]string, len(chans))

	for i, cha := range chans {
		channelIds[i] = cha["id"].(string)
	}

	return channelIds, nil
}

func GetChannelMessagesSinceLastRun(token, channel string, lastRunTime string)([]map[string]interface{}, error) {

	cachedToken = token

	method := fmt.Sprintf("%schannels.history?token=%s&channel=%s&oldest=%s", baseUrl, token, channel, lastRunTime)

	response, err := simhttp.GetResponseAsMap(method)

	if(err != nil || response == nil){return nil, err}

	messages := make([]map[string]interface{}, len(response["messages"].([]interface{})))

	for i, message := range response["messages"].([]interface{}) {
		messages[i] = message.(map[string]interface{})}

	return messages, nil
}

func PostMessageToDefaultChannel(message string)(map[string]interface{}, error) {
	// todo: do some testing to determine string escaping needs.
	// fmt.Sprintf("chat.postMessage?token=%s&username=ChOPS&channel=%s&text="
	return simhttp.GetResponseAsMap(baseUrl + message)
}

func GetUserInfo(token, userId string)(map[string]interface{}, error){

	cachedToken = token

	method := fmt.Sprintf("%susers.info?token=%s&user=%s", baseUrl, token, userId)

	response, err := simhttp.GetResponseAsMap(method)

	if err != nil || response == nil {return nil, err}

	userInfo := response["user"].(map[string]interface{})

	return userInfo, nil
}

func GetUserInfoWithCachedToken(userId string)(map[string]interface{}, error) {

	return GetUserInfo(cachedToken, userId)
}