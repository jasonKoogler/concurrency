package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Result represents a positive search result, it will be the name of the file
type Result string

// ResultMsg represents the message passed to through the channel, it will contain
// either a [Data] string or an [Err] error
type ResultMsg struct {
	Data Result
	Err  error
}

// NewResultMsg is a constructor function for creating a ResultMsg more easily
// the data parameter will be caste to a Result to simplify usage.
func NewResultMsg(data string, err error) *ResultMsg {
	return &ResultMsg{
		Data: Result(data),
		Err:  err,
	}
}

// JobManager is the central struct which provide various concurrency primitives
type JobManager struct {
	wg          *sync.WaitGroup
	resultmsgch chan *ResultMsg
	result      []*Result
	readDone    *sync.WaitGroup
}

// Newworker create a new instance of the JobManager object
func NewJobManager() *JobManager {
	return &JobManager{
		wg:          new(sync.WaitGroup),
		resultmsgch: make(chan *ResultMsg),
		result:      []*Result{},
		readDone:    new(sync.WaitGroup),
	}
}

// SearchFiles concurrently searches for files within a directory, well not really
// it  requires the files to be provided as a slice of strings.  Further higher level functions
// will make handling various use cases more easily.
//
// todo: create higher level functions that grab files from dirs, dirs in dirs, or a list of dirs
// ---- then call search files
func (jm *JobManager) SearchFiles(term string, files ...fs.DirEntry) []*Result {
	jm.readDone.Add(1)

	go jm.searchWorkerReceive()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		jm.wg.Add(1)
		go jm.searchWorkerSend(file.Name(), term)
	}
	jm.wg.Wait()
	close(jm.resultmsgch)
	jm.readDone.Wait()

	return jm.result
}

// serachWorkerSend takes a file and a term as a string and searches the file for the string
// it sends either an error or a result to the resultmsgch
func (jm *JobManager) searchWorkerSend(file string, term string) {
	defer jm.wg.Done()
	b, err := os.ReadFile(file)
	if err != nil {
		jm.resultmsgch <- NewResultMsg("", err)
		return
	}
	if byteToStringSearchForTerm(b, term) {
		jm.resultmsgch <- NewResultMsg(file, nil)
	}
}

func (jm *JobManager) searchWorkerReceive() {
	defer jm.readDone.Done()
	for msg := range jm.resultmsgch {
		if msg.Err != nil {
			log.Println("Error reading file: ", msg.Err)
		} else {
			jm.result = append(jm.result, &msg.Data)
		}
	}
}

// byteToStringSearchForTerm converts bytes to string and then searches for a match
func byteToStringSearchForTerm(b []byte, term string) bool {
	content := string(b)
	return strings.Contains(content, term)
}

func main() {

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting the pwd: %s", err)
	}
	workingFilesPath := filepath.Join(pwd, "files")
	files, err := os.ReadDir(workingFilesPath)
	if err != nil {
		log.Fatalf("error reading dir: %s", err)
	}

	if err := os.Chdir(workingFilesPath); err != nil {
		log.Fatalf("error changing working directory: %s", err)
	}

	jm := NewJobManager()

	jm.SearchFiles("the", files...)

	os.Chdir(pwd)

	for _, addr := range jm.result {
		fmt.Println(*addr)
	}
}
