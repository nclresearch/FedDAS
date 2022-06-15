package request

import (
	"sync"
	"time"
)

// This Go file contains all of the global variables for the API.

/*
	The RequestTask struct.
	When you make your request, this is what it will return.
	This contains a summary of the request. Altering the contents of the struct does not alter the request.
*/
type RequestTask struct {
	AgentId string
	API     string
	Command string
	Time    time.Time
	Args    interface{}
	Timeout float64
	/*
		This is specifically make into a channel.
		If you need to get the result of a request, receive from this channel like so.
		result := <- task.Result
		Note that this will lock your current thread to sleep. Timeouts and Errors will also be reported here.
	*/
	Result chan Result
}

/*
	The RequestTask object for internal use
	This is to prevent accidental alteration to the request that would break the controller
	This includes an extra Ack flag.
	The Args will be called and used when processing responses
*/
type ImplRequestTask struct {
	AgentId string
	API     string
	Command string
	Ack     bool
	Time    time.Time
	Args    interface{}
	Timeout float64
}

/*
	The result contents of an API request
*/
type Result struct {
	ResultType ResultType
	/*
		Depending on the ResultType, the Content can be anything from nil, error, to some other return type.
		Please refer to the documentation in GitHub
	*/
	Content interface{}
}

type ResultType string

const (
	Error ResultType = "error"
	Ok    ResultType = "ok"
)

type RegisterRequest struct {
	ID            string   `json:"requestId"`
	Direction     string   `json:"direction"`
	Status        string   `json:"status"`
	InternalIP    string   `json:"internalIP"`
	DeviceSupport []string `json:"devSupport"`
	SensorSupport []string `json:"sensorSupport"`
}

/*
	The request struct
	Used to represent a request's contents
*/
type Request struct {
	RequestID string                 `json:"requestId"`
	Command   string                 `json:"command"`
	Args      map[string]interface{} `json:"args"`
}

//This stores all ongoing requests to the agent.
/*
	Request map that temporarily stores ongoing requests.
	Request ID -> Request struct
	Each entry will be removed once it timeout or completed.
	Since this map is written by several goroutines. This needs to be synchronized.
*/
var DeployRequests sync.Map
var MonitorRequests sync.Map

//This stores all request tasks objects.
var DeployTaskList sync.Map
var MonitorTaskList sync.Map
