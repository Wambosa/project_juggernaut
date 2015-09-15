package simhttp

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
)

const Version  = 1


func GetResponseAsString(url string) (string, error) {

	resp, err := http.Get(url)

	if err != nil {return "", err}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {return "", err}

	return string(buf), err
}

func GetResponseAsMap(url string) (map[string]interface{}, error) {

	raw, err := GetResponseAsString(url)

	if err != nil { return nil, err}

	var newMap map[string]interface{}

	json.Unmarshal([]byte(raw), &newMap)

	return newMap, err
}