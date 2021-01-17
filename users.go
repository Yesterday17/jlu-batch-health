package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

// sync map[int64]*User
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
		Users.Store(u.Username, u)
	}
	return nil
}
