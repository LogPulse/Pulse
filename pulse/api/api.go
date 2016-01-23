package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gophergala2016/Pulse/pulse"
	"github.com/gophergala2016/Pulse/pulse/email"
	"github.com/gophergala2016/Pulse/pulse/file"
)

// Result : is used for ResponseWriter in handlers
type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var buffStrings []string

// Start : will start the REST API
func Start() {
	http.HandleFunc("/log/message", StreamLog)
	http.HandleFunc("/log/file", SendFile)
	http.ListenAndServe(":8080", nil)
}

// StreamLog : Post log statement to our API
func StreamLog(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	decoder := json.NewDecoder(r.Body)
	var body struct {
		Message string `json:"message"`
	}

	err := decoder.Decode(&body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(Result{200, "success"})
	io.WriteString(w, string(result))

}

// SendFile : Post log files to our API
func SendFile(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	f, header, err := r.FormFile("file")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	defer f.Close()

	decoder := json.NewDecoder(r.Body)
	var body struct {
		Email string `json:"email"`
	}

	err = decoder.Decode(&body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	extension := filepath.Ext(header.Filename)
	filename := header.Filename[0 : len(header.Filename)-len(extension)]

	stdIn := make(chan string)
	email.ByPassMail = true // Needs to bypass emails and store in JSON
	email.OutputFile = filename + ".json"
	email.EmailList = []string{body.Email}
	pulse.Run(stdIn, email.Send)

	line := make(chan string)
	file.StreamRead(f, line)
	for l := range line {
		if l == "EOF" {
			email.ByPassMail = false
			// Once EOF, time to send email from cache JSON storage
			go email.SendFromCache(email.OutputFile)
			continue
		}
		stdIn <- l
	}
	close(stdIn)

	fmt.Fprintf(w, "Pattern found and saving!")
}
