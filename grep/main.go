package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Result represents a positive search result, it will be the name of the file
type Result struct {
	Filename   string
	LineNumber int
	Text       string
}

// NewResult helper function for creating a new result object
func NewResult(fileName string, lineNumber int, text string) Result {
	return Result{
		Filename:   fileName,
		LineNumber: lineNumber,
		Text:       text,
	}
}

// ResultMsg represents the message passed to through the channel, it will contain
// either a [Data] string or an [Err] error
type ResultMsg struct {
	Data []Result
	Err  error
}

// NewResultMsg is a constructor function for creating a ResultMsg more easily
// the data parameter will be caste to a Result to simplify usage.
func NewResultMsg(data []Result, err error) *ResultMsg {
	return &ResultMsg{
		Data: data,
		Err:  err,
	}
}

// JobManager is the central struct which provide various concurrency primitives
type JobManager struct {
	wg          *sync.WaitGroup
	resultmsgch chan *ResultMsg
	result      *[]Result
	readDone    *sync.WaitGroup
}

// Newworker create a new instance of the JobManager object
func NewJobManager(numFiles int) *JobManager {
	var bufferSize int
	if numFiles >= 3 {
		bufferSize = 5
	} else if numFiles > 50 && numFiles <= 1000 {
		bufferSize = 10
	} else if numFiles > 1000 {
		bufferSize = 100
	}

	return &JobManager{
		wg:          new(sync.WaitGroup),
		resultmsgch: make(chan *ResultMsg, bufferSize),
		result:      new([]Result),
		readDone:    new(sync.WaitGroup),
	}
}

// SearchFiles concurrently searches for files within a directory, well not really
// it  requires the files to be provided as a slice of strings.  Further higher level functions
// will make handling various use cases more easily.
//
// todo: create higher level functions that grab files from dirs, dirs in dirs, or a list of dirs
// ---- then call search files
func (jm *JobManager) SearchFiles(term string, files ...fs.DirEntry) *[]Result {
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
	f, err := os.Open(file)
	if err != nil {
		jm.resultmsgch <- NewResultMsg(nil, err)
		return
	}
	defer f.Close()

	jm.resultmsgch <- scanByLine(f, term)
}

// searchWorkerReceive receives values from the resultmsg channel
// so the channel wont block further goroutine access
// !! should be run in its own goroutine
func (jm *JobManager) searchWorkerReceive() {
	defer jm.readDone.Done()
	for msg := range jm.resultmsgch {
		if msg.Err != nil {
			log.Println("Error reading file: ", msg.Err)
		} else {
			*jm.result = append(*jm.result, msg.Data...)
		}
	}
}

// scanByLine takes an os.File and a term to search for and
// scans the file line by line for the term, adding all
// positive results to the ResultMsg as well as errors
func scanByLine(file *os.File, term string) *ResultMsg {
	scanner := bufio.NewScanner(file)
	lineNum := 1
	resultmsg := new(ResultMsg)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, term) {
			resultmsg.Data = append(resultmsg.Data, NewResult(file.Name(), lineNum, term))
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		resultmsg.Err = err
		return resultmsg
	}

	return resultmsg
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

	jm := NewJobManager(len(files))

	jm.SearchFiles("and", files...)

	os.Chdir(pwd)

	for _, found := range *jm.result {
		fmt.Println(found)
	}

}
