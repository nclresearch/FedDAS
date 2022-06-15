package request

import (
	"encoding/json"
	"github.com/lithammer/shortuuid"
	"github.com/streadway/amqp"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"osmoticframework/controller/types"
	"time"
)

/*
All of the deploy API request functions.
It constructs the request body and publishes it to the queue. For information on the structure. See deploying API in the wiki
An internal request ID will be automatically generated upon calling the function.
Each api request has their own timeout duration. If you think it needs more time, please change the value.
*/
func RunRequest(agentId string, deployArgs types.DeployArgs, authInfo types.AuthInfo, timeout float64) *RequestTask {
	id := shortuuid.New()
	for {
		id = shortuuid.New()
		if r, _ := DeployRequests.Load(id); r == nil {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "run",
		Args: map[string]interface{}{
			"deployArgs": deployArgs,
			"authInfo":   authInfo,
		},
	})
	log.Info.Printf("%s << Deploy request with image %s\n", agentId, deployArgs.Image)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "run",
		Time:    time.Now(),
		Args: map[string]interface{}{
			"deployArgs": deployArgs,
			"authInfo":   authInfo,
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	DeployTaskList.Store(id, task)
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "run",
		Ack:     false,
		Time:    time.Now(),
		Timeout: timeout,
	})
	return &task
}

func StopRequest(agentId, containerId string, timeout float64) *RequestTask {
	var id string
	for {
		id := shortuuid.New()
		if _, ok := DeployRequests.Load(id); !ok {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "stop",
		Args: map[string]interface{}{
			"containerId": containerId,
		},
	})
	log.Info.Printf("%s << Stop request on container %s\n", agentId, containerId)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "stop",
		Ack:     false,
		Time:    time.Now(),
		Args: map[string]string{
			"containerId": containerId,
		},
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "stop",
		Time:    time.Now(),
		Args: map[string]string{
			"containerId": containerId,
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	DeployTaskList.Store(id, task)
	return &task
}

func DeleteRequest(agentId, containerId string, deleteImage bool, timeout float64) *RequestTask {
	var id string
	for {
		id := shortuuid.New()
		if _, ok := DeployRequests.Load(id); !ok {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "delete",
		Args: map[string]interface{}{
			"containerId": containerId,
			"deleteImage": deleteImage,
		},
	})
	log.Info.Printf("%s << Delete request on container %s\n", agentId, containerId)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "delete",
		Ack:     false,
		Time:    time.Now(),
		Args: map[string]string{
			"containerId": containerId,
		},
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "delete",
		Time:    time.Now(),
		Args: map[string]interface{}{
			"containerId": containerId,
			"deleteImage": deleteImage,
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	DeployTaskList.Store(id, task)
	return &task
}

func UpdateRequest(agentId, containerId string, deployArgs types.DeployArgs, authInfo types.AuthInfo, timeout float64) *RequestTask {
	var id string
	for {
		id := shortuuid.New()
		if _, ok := DeployRequests.Load(id); !ok {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "update",
		Args: map[string]interface{}{
			"containerId": containerId,
			"deployArgs":  deployArgs,
			"authInfo":    authInfo,
		},
	})
	log.Info.Printf("%s << Update request on container %s\n", agentId, containerId)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "update",
		Ack:     false,
		Time:    time.Now(),
		//We need to store the old container ID so that we can replace the entry in the database
		Args: map[string]string{
			"oldContainerId": containerId,
		},
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "update",
		Time:    time.Now(),
		Args: map[string]interface{}{
			"containerId": containerId,
			"deployArgs":  deployArgs,
			"authInfo":    authInfo,
		},
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	DeployTaskList.Store(id, task)
	return &task
}

func ListRequest(agentId string, timeout float64) *RequestTask {
	var id string
	for {
		id := shortuuid.New()
		if _, ok := DeployRequests.Load(id); !ok {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "list",
		Args:      map[string]interface{}{},
	})
	log.Info.Printf("%s << List request\n", agentId)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "list",
		Ack:     false,
		Time:    time.Now(),
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "list",
		Time:    time.Now(),
		Args:    nil,
		Timeout: timeout,
		Result:  make(chan Result, 1),
	}
	DeployTaskList.Store(id, task)
	return &task
}

func InspectRequest(agentId, containerId string, timeout float64) *RequestTask {
	var id string
	for {
		id := shortuuid.New()
		if _, ok := DeployRequests.Load(id); !ok {
			break
		}
	}
	request, _ := json.Marshal(Request{
		RequestID: id,
		Command:   "inspect",
		Args: map[string]interface{}{
			"containerId": containerId,
		},
	})
	log.Info.Printf("%s << Inspect container on container %s\n", agentId, containerId)
	err := queue.Ch.Publish(
		"",
		"deploy-"+agentId,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        request,
		},
	)
	if err != nil {
		log.Error.Println("Failed sending request")
		log.Error.Println(err)
		return nil
	}
	DeployRequests.Store(id, ImplRequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "inspect",
		Ack:     false,
		Time:    time.Now(),
		Args: map[string]string{
			"containerId": containerId,
		},
		Timeout: timeout,
	})
	task := RequestTask{
		AgentId: agentId,
		API:     "deploy",
		Command: "inspect",
		Time:    time.Now(),
		Args: map[string]string{
			"containerId": containerId,
		},
		Timeout: timeout,
		Result:  nil,
	}
	DeployTaskList.Store(id, task)
	return &task
}
