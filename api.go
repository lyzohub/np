package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// CmdExecRequest represents request to execute a command for a given ID
type CmdExecRequest struct {
	Cmd string `json:"cmd"`
	ID  int    `json:"id"`
}

// CmdExecResult is a result of command execution for a given ID
type CmdExecResult struct {
	Cmd    string `json:"cmd"`
	ID     int    `json:"id"`
	Result string `json:"result"`
}

// APIError defines common format for API errors
type APIError struct {
	Error string `json:"error"`
}

var (
	cmdMutexMap sync.Map
)

func getNamedMutex(key string) *sync.Mutex {
	mutex, _ := cmdMutexMap.LoadOrStore(key, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, APIError{Error: "Method not allowed"}, http.StatusInternalServerError)
		return
	}

	// Limit the size of the request body to 1MB.
	// Such big request body may indicate some malicious activity.
	// NOTE: In real project I would extract this to the middleware.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, APIError{Error: "Error reading request body."}, http.StatusInternalServerError)
		return
	}

	var req CmdExecRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSON(w, APIError{Error: "Invalid writeJSON format."}, http.StatusBadRequest)
		return
	}

	if req.Cmd == "" || req.ID == 0 {
		writeJSON(w, APIError{Error: "Invalid writeJSON: 'cmd' and 'id' are required."}, http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%d+%s", req.ID, req.Cmd)

	// If same command execution is in progress, don't start a duplicated one, request should wait and return result.
	mutex := getNamedMutex(key)
	mutex.Lock()
	defer mutex.Unlock()
	defer cmdMutexMap.Delete(key)

	result, exists := database.Get(key)
	if !exists {
		result = ExecuteCommand(req.Cmd)
		err = database.Set(key, result)
		if err != nil {
			writeJSON(w, APIError{Error: "Error saving result."}, http.StatusInternalServerError)
			return
		}
	}

	resp := CmdExecResult{
		Cmd:    req.Cmd,
		ID:     req.ID,
		Result: result,
	}
	writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(jsonData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ExecuteCommand mimics real command execution
func ExecuteCommand(command string) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Result of command is random string
	fmt.Printf("Executing command: %s\n", command)
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	// Mimic real execution by sleeping between 100ms and 1000ms
	time.Sleep(time.Duration(rand.Intn(900)+100) * time.Millisecond)
	fmt.Printf("Command %s was executed successfully\n", command)
	return string(b)
}
