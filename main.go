package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {

}

func InitConfig() (*config.Configuration, error) {
	var conf config.Configuration

	data, err := ioutil.ReadFile("./config/keys.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println("error: ", err)
		return nil, err
	}

	fmt.Println(configParam.BindAddr)
	return conf, nil
}
