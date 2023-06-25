package db

import (
	"fmt"

	"github.com/google/uuid"
)

var usersSession = UsersTemplate{}
var fileSession = FileTemplate{}

// map[username]Session
var activeSession = make(map[string]Session)

func NewUUID() string {
	id, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
	}

	return id.String()
}
