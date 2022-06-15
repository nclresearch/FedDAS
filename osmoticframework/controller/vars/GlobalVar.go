package vars

import (
	"os"
	"osmoticframework/controller/log"
	"sync"
)

//Global variables

//Stores all agents and containers registered in memory
var Agents sync.Map

//Stores all deployed cloud resources (Deployments, Service, Jobs, Cronjobs)
//Each map stores the name of the resource in its key, value is always nil and ignored.
var Deployments sync.Map
var Services sync.Map
var Jobs sync.Map
var CronJobs sync.Map

//Ready flag
//Since the listener takes some time to get ready, you can use this flag to wait for the listener to get ready
//This value will be set to true when it starts listening to responses
//Call GetApiReady() to get its status
var apiReady = false

//Terminate flag
var terminate = false

func SetApiReady() {
	apiReady = true
}

func GetApiReady() bool {
	return apiReady
}

func SetTerminate() {
	terminate = true
}

func IsTerminate() bool {
	return terminate
}

func GetCredDirectory() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal.Println("Unable to obtain current directory")
		log.Fatal.Panicln(err)
	}
	logsDir := pwd + "/orchestrator-cred"
	stat, err := os.Stat(logsDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(logsDir, 0700)
		log.Info.Println("Controller credential directory created")
	} else if stat.Mode() != (0700 | os.ModeDir) {
		log.Fatal.Fatalln("Credential directory insecure. Must be 0700")
	}
	if err != nil {
		log.Fatal.Println("failed writing to current directory")
		log.Fatal.Panicln(err)
	}
	return logsDir
}
