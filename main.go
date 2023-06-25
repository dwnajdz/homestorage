package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"homestorage/db"
)

func main() {
	// page routes
	http.HandleFunc("/", index)
	http.HandleFunc("/proxy", proxy)
	http.HandleFunc("/import", importFile)
	http.HandleFunc("/library", library)
	http.HandleFunc("/download", downloadToClient)
	http.HandleFunc("/info/", fileInfo)
	// user interface
	http.HandleFunc("/login", login)
	http.HandleFunc("/signup", signUp)
	http.HandleFunc("/signout", signOut)

	log.Println("listening on http://localhost:80")
	http.ListenAndServe(":80", nil)
}

type BaseTemplate struct {
	Title  string
	Token  string
	UserID string
	Files  []db.File
}

func getToken(r *http.Request) db.Session {
	token := r.URL.Query().Get("token")
	return db.DecodeSession(token)
}

func tokenVerify(token string, w http.ResponseWriter, r *http.Request) db.Session {
	if len(token) <= 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	session := db.DecodeSession(token)
	if !session.VerifySession(r) {
		http.Redirect(w, r, "/proxy?token="+token, http.StatusFound)
	}
	if !session.IsLogged {
		http.Redirect(w, r, "/login?status=nopermission", http.StatusFound)
	}

	return session
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").ParseFiles("frontend/index.html", "frontend/base.html")
	if err != nil {
		log.Println(err)
	}

	token := r.URL.Query().Get("token")
	session := tokenVerify(token, w, r)
	files := db.FilesQuery(session.CurrentUser.ID)

	data := BaseTemplate{
		Title: "Overview",
		Token: token,
		Files: files,
	}

	if err = tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println(err)
	}
}

func library(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").ParseFiles("frontend/library.html", "frontend/base.html")
	if err != nil {
		log.Println(err)
	}

	token := r.URL.Query().Get("token")
	session := tokenVerify(token, w, r)
	files := db.FilesQuery(session.CurrentUser.ID)

	data := BaseTemplate{
		Title:  "Library",
		Token:  token,
		UserID: session.CurrentUser.ID,
		Files:  files,
	}

	if err = tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println(err)
	}
}

// FILES IMPORT/DOWNLOAD
func ifStorageExist() {
	if _, err := os.Stat("storage"); os.IsNotExist(err) {
		err = os.Mkdir("storage", 0755)
		if err != nil {
			fmt.Println("Could not create storage directory.")
			fmt.Println(err)
		}
	} else {
		return
	}
}

func getUserPath(token string) string {
	session := db.DecodeSession(token)
	userPath := fmt.Sprintf("storage/%s/", session.CurrentUser.ID)
	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		err = os.Mkdir(userPath, 0755)
		if err != nil {
			fmt.Println("Could not create user directory.")
			fmt.Println(err)
		}
	}
	return userPath
}

func importFile(w http.ResponseWriter, r *http.Request) {
	// verify user token
	token := r.URL.Query().Get("token")
	session := tokenVerify(token, w, r)
	// if some of the storages does not exist create new
	ifStorageExist()
	userPath := getUserPath(token)
	r.ParseMultipartForm(100)

	file, handler, err := r.FormFile("importFile")
	if err != nil {
		fmt.Println("Error Retrieving the File.")
		fmt.Println(err)
	}
	defer file.Close()
	// add file to database
	db.FileAdd(session.CurrentUser.ID, handler.Filename, handler.Size, handler.Header)

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	if err = os.WriteFile(fmt.Sprintf("%s%s", userPath, handler.Filename), fileBytes, 0644); err != nil {
		fmt.Println("Failed creating a file.")
		fmt.Println(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/?token=%s", token), http.StatusFound)
}

func downloadToClient(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	fileUrl := r.URL.Query().Get("file")
	if fileUrl == "" {
		// return 400 HTTP response code for BAD REQUEST
		http.Error(w, "Update failed due to malformed URL.", 400)
		return
	}
	fileName := fmt.Sprintf("%s%s", getUserPath(token), fileUrl)

	file, err := os.Open(fileName)
	if err != nil {
		// return 404 HTTP response code for File not found
		http.Error(w, "file not found.", 404)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	//Transmit the headers
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileUrl)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	io.Copy(w, file) // transmit the updatefile bytes to the client
}

type FileInfoData struct {
	Title  string
	Token  string
	File   db.File
	SizeMB float64
}

func fileInfo(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").ParseFiles("frontend/info.html", "frontend/base.html")
	if err != nil {
		log.Println(err)
	}

	token := r.URL.Query().Get("token")
	session := tokenVerify(token, w, r)
	filename := strings.Split(r.URL.Path, "/")[2]
	file := db.FileQuery(session.CurrentUser.ID, filename)

	base := math.Log(float64(file.Size)) / math.Log(1024)
	sizeMB := math.Pow(1024, base-math.Floor(base))

	data := FileInfoData{
		Title:  "Info",
		Token:  token,
		File:   file,
		SizeMB: sizeMB,
	}

	if err = tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println(err)
	}
}

// USER AUTHENTICATION
func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("frontend/login.html"))

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		current_user := db.UsersQuery(email)
		if current_user.UserCheckPassword(password) {
			swt_session := db.NewSWTSession(current_user, r)
			http.Redirect(w, r, fmt.Sprintf("/?token=%s", swt_session), http.StatusFound)
		}
	}

	tmpl.Execute(w, nil)
}

func signUp(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("frontend/signup.html"))

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		id := db.NewUUID()

		user := db.User{
			ID:       id,
			Email:    email,
			Username: username,
			Password: db.HashPassword(password),
			IsAdmin:  true,
		}

		if !db.DoesThisUserExist(email) {
			db.UsersAdd(email, user)
			http.Redirect(w, r, "/login?status=success", http.StatusFound)
		} else {
			http.Redirect(w, r, "/signup?status=user already exist", http.StatusFound)

		}
	}

	tmpl.Execute(w, nil)
}

func signOut(w http.ResponseWriter, r *http.Request) {
	session := getToken(r)
	session.SignOut()

	http.Redirect(w, r, "/login?status=signout", http.StatusFound)
}

func proxy(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("frontend/proxy.html"))
	session := getToken(r)

	if r.Method == "POST" {
		password := r.FormValue("password")

		current_user := db.UsersQuery(session.CurrentUser.Email)
		if current_user.UserCheckPassword(password) {
			session.SignOut()
			swt_session := db.NewSWTSession(current_user, r)
			http.Redirect(w, r, fmt.Sprintf("/?token=%s", swt_session), http.StatusFound)
		} else {
			http.Redirect(w, r, "/login?status=fail", http.StatusFound)
		}
	}

	tmpl.Execute(w, nil)
}
