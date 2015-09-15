package conf

import (
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"os"
)

const Version = 1

func getThisFolder() string{

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	return dir
}

func LoadRaw(path string) (string, error) {

	buf, err := ioutil.ReadFile(path)

	return string(buf), err
}

func LoadJson(path string) (map[string]interface{}, error){

	buf, err := ioutil.ReadFile(path)

	if err != nil {return nil, err}

	var newMap map[string]interface{}

	json.Unmarshal(buf, &newMap)

	return newMap, err
}

func SaveJson(path string, alteredConfig map[string]interface{}) error {

	buf, err := json.MarshalIndent(alteredConfig, "", "  ")

	if err != nil {return err}

	ioutil.WriteFile(path, buf, 0644)

	return nil
}

func LoadSqlite(query string) (map[string]string, error) {

	//todo: load from the sql database instead of a file
	return nil, nil
}

func SaveSqlite(query string) error {
	//todo: later just use polyjug instead. (query string databaseName string)
	return nil
}