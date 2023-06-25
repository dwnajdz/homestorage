package db

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Email    string
	Username string
	Password string
	IsAdmin  bool
}

// USERS
func UsersAdd(key string, value User) {
	usersSession.Mu.Lock()
	defer usersSession.Mu.Unlock()
	if len(usersSession.Cargo) == 0 {
		m := make(map[string]User)
		m[key] = value
		usersSession.Cargo = m

		return
	}
	usersSession.Cargo[key] = value
}

func UsersDelete(key string) {
	usersSession.Mu.Lock()
	defer usersSession.Mu.Unlock()

	delete(usersSession.Cargo, key)
}

func UsersQuery(key string) User {
	return usersSession.Cargo[key]
}

func DoesThisUserExist(key string) bool {
	_, ok := usersSession.Cargo[key]
	return ok
}

func (user User) UserCheckPassword(password string) bool {
	return CheckPasswordHash(password, user.Password)
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(bytes)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
