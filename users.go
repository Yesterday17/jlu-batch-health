package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

var Users sync.Map

func LoadUsers() error {
	data, err := ioutil.ReadFile(Config.AccountsPath)
	if err != nil {
		return err
	}

	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.Fields == nil {
			u.Fields = map[string]string{}
		}
		u.Jar = NewHealthJar()
		Users.Store(u.Username, &u)
	}
	return nil
}
