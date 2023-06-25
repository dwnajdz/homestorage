package db

import (
	"net/textproto"
	"sync"
	"time"
)

type FileTemplate struct {
	Mu sync.Mutex
	// map[userid]map[filename]File
	Cargo map[string][]File
}

type File struct {
	Name   string
	Owner  string
	Size   int64
	Header textproto.MIMEHeader
	Added  time.Time
}

// FILES
func FileAdd(userid, filename string, size int64, header textproto.MIMEHeader) {
	fileSession.Mu.Lock()
	defer fileSession.Mu.Unlock()

	file := File{
		Name:   filename,
		Owner:  userid,
		Size:   size,
		Header: header,
		Added:  time.Now(),
	}

	cargo := fileSession.Cargo[userid]
	cargo = append(cargo, file)

	if len(fileSession.Cargo) <= 0 {
		m := make(map[string][]File)
		m[userid] = append(m[userid], file)
		fileSession.Cargo = m

		return
	}

	fileSession.Cargo[userid] = cargo
}

func FileDelete(key string) {
	fileSession.Mu.Lock()
	defer fileSession.Mu.Unlock()

	delete(fileSession.Cargo, key)
}

//func FileQuery(userid, filename string) File {
//	return fileSession.Cargo[userid][filename]
//}

func FilesQuery(userid string) []File {
	m, ok := fileSession.Cargo[userid]
	if ok {
		return m
	}
	return nil
}

func FileQuery(userid, filename string) File {
	files, ok := fileSession.Cargo[userid]
	if ok {
		for _, file := range files {
			if file.Name == filename {
				return file
			}
		}
	}
	return File{}
}

func DoesThisFileExist(key string) bool {
	_, ok := fileSession.Cargo[key]
	return ok
}
