package db

import (
	"net/http"

	"github.com/wspirrat/swt/swt"
)

type Session struct {
	CurrentUser User
	IsLogged    bool
	UserIP      string
}

func NewSWTSession(user User, r *http.Request) string {
	session := Session{
		CurrentUser: user,
		IsLogged:    true,
		UserIP:      ReadUserIP(r),
	}
	activeSession[user.Email] = session

	return swt.EncodeSWT(session)
}

func DecodeSession(payload string) Session {
	decoded := swt.DecodeSWT(payload)

	if !decoded.IsPayloadNil() {
		//token := decoded.Payload.(Session)
		//querysession, ok := activeSession[token.CurrentUser.Email]
		//fmt.Println("qs:", querysession, ok)
		//fmt.Println("decoded:", decoded)
		//if token == querysession {
		//	return token
		//} else if !ok {
		//	return token
		//}
		return decoded.Payload.(Session)
	}

	return Session{}
}

func (session Session) VerifySession(r *http.Request) bool {
	if _, ok := activeSession[session.CurrentUser.Email]; ok {
		if session.UserIP == ReadUserIP(r) {
			return true
		}
		delete(activeSession, session.CurrentUser.Email)
	}
	return false
}

func (session Session) SignOut() {
	delete(activeSession, session.CurrentUser.Email)
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")

	//if IPAddress == "" {
	//	IPAddress = r.Header.Get("X-Forwarded-For")
	//}
	//if IPAddress == "" {
	//	IPAddress = r.RemoteAddr
	//}
	return IPAddress
}
