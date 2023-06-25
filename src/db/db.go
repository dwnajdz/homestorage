package db

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type FileTemplate struct {
	Mu sync.Mutex
	// map[userid]map[filename]File
	Cargo map[string][]File
}

type UsersTemplate struct {
	Mu    sync.Mutex
	Cargo map[string]User
}

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
