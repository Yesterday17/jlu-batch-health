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
		user := u
		if user.Fields == nil {
			user.Fields = map[string]string{}
		}
		user.Jar = NewHealthJar()
		Users.Store(u.Username, &user)
	}
	return nil
}
