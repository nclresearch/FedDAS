package callback

import (
	"errors"
	"github.com/mitchellh/mapstructure"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/database"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
)

//Reads the response message from the agent.
//See README.md for the message structure
func ParseDeploy(message map[string]interface{}) {
	requestId, ok := message["requestId"].(string)
	if !ok {
		//Cannot figure out what the request is. Ignore completely.
		return
	}
	_requestTask, exists := request.DeployRequests.Load(requestId)
	if !exists || requestId == "internal" {
		//Ignore all requests if the request does not exist in memory. This can occur if the controller just crashed and recovered
		//It is also possible if agents reply after the request has been timed out. In that case, you should increase the timeout duration.
		//Request ID 'internal' refers to actions that are performed by the agent itself, without any interaction from the controller. Thus, they will be ignored
		return
	}
	//The message itself only contains the request ID. It does not have any info on what the request to
	//Here we refer the message request ID to the request ID stored in the controller, where it contains information such as the api, command, arguments, etc.
	requestTask := _requestTask.(request.ImplRequestTask)
	command := requestTask.Command
	status, ok := message["status"].(string)
	if !ok {
		//No response status. Ignore
		return
	}
	agentId := requestTask.AgentId
	switch status {
	//If the agent returns an ack.
	case "ack":
		//Update the ack flag. So that the controller will not clean up this request since the request has been acknowledged
		//Since the map is written between different goroutines. The whole Request struct needs to be reassigned instead of editing the element through pointers.
		ongoingRequest := request.ImplRequestTask{
			AgentId: agentId,
			API:     requestTask.API,
			Command: command,
			Ack:     true,
			Time:    requestTask.Time,
			Args:    requestTask.Args,
			Timeout: requestTask.Timeout,
		}
		request.DeployRequests.Store(requestId, ongoingRequest)
	//The agent completes the command and returns ok
	case "ok":
		switch command {
		case "run":
			containerId, ok := message["containerId"].(string)
			if !ok {
				err := errors.New("invalid response")
				log.Error.Printf("%s (req: %s) >> Failed to parse new container deployed\n", agentId, requestId)
				log.Error.Println(err)
				CallbackError(requestId, err)
				return
			}
			database.AddContainer(agentId, containerId)
			CallbackOk(requestId, containerId)
			log.Info.Printf("%s (req: %s) >> Container %s started\n", agentId, requestId, containerId)
		case "stop":
			//We have the container ID stored in memory. No need to parse the message
			containerId := requestTask.Args.(map[string]string)["containerId"]
			database.StopContainer(agentId, containerId)
			log.Info.Printf("%s (req: %s) >> Container %s stopped\n", agentId, requestId, containerId)
			CallbackOk(requestId, nil)
		case "delete":
			//We have the container ID stored in memory. No need to parse the message
			containerId := requestTask.Args.(map[string]string)["containerId"]
			database.RemoveContainer(agentId, containerId)
			CallbackOk(requestId, nil)
			log.Info.Printf("%s (req: %s) >> Container %s deleted\n", agentId, requestId, containerId)
		case "update":
			//We have the old container ID stored in memory.
			//We can get the new container ID from the agent's response
			//With these two we can update the database

			//Get the old container ID from memory
			containerId := requestTask.Args.(map[string]string)["oldContainerId"]
			//Parse the agent's response to get the new container ID
			newContainerId, ok := message["containerId"].(string)
			if !ok {
				err := errors.New("invalid response")
				log.Error.Printf("%s (req: %s) >> Failed to parse new container from update\n", agentId, requestId)
				log.Error.Println(err)
				CallbackError(requestId, err)
				return
			}
			//Update the database
			database.UpdateContainer(agentId, containerId, newContainerId)
			CallbackOk(requestId, newContainerId)
			log.Info.Printf("%s (req: %s) >> Container %s updated to %s\n", agentId, requestId, containerId, newContainerId)
		case "list":
			//Agent sends an array of container structs. Decoding is required.
			var containers []types.Container
			err := mapstructure.Decode(message["containers"], &containers)
			if err != nil {
				//Broken response. Or decoding failed
				log.Error.Printf("%s (req: %s) >> Failed to decode container listing\n", agentId, requestId)
				log.Error.Println(err)
				return
			}
			CallbackOk(requestId, containers)
			log.Info.Printf("%s (req: %s) >> Received container listing\n", agentId, requestId)
		case "inspect":
			//Agent sends a container object. Requires decode
			var container types.Container
			err := mapstructure.Decode(message["container"], &container)
			if err != nil {
				//Broken response. Or decoding failed
				log.Error.Printf("%s (req: %s) >> Failed to decode container\n", agentId, requestId)
				log.Error.Println(err)
				return
			}
			CallbackOk(requestId, container)
			log.Info.Printf("%s (req: %s) << Received container inspection\n", agentId, requestId)
		case "prom":
			containerId, ok := message["containerId"].(string)
			if !ok {
				err := errors.New("invalid response")
				log.Error.Printf("%s (req: %s) >> Failed to parse new container deployed\n", agentId, requestId)
				log.Error.Println(err)
				CallbackError(requestId, err)
				return
			}
			//We don't need to write down the container ID for Prometheus
			CallbackOk(requestId, containerId)
			log.Info.Printf("%s (req: %s) >> Prometheus container %s started", agentId, requestId, containerId)
		default:
			//Ignore message
			return
		}
	//If the request failed.
	case "failed":
		//Find which request failed in memory.
		var err error
		errStr, ok := message["error"].(string)
		if !ok {
			err = errors.New("unknown error")
		} else {
			err = errors.New(errStr)
		}
		log.Error.Printf("%s (req: %s) >> Request failed", requestTask.AgentId, requestId)
		log.Error.Println(err)
		CallbackError(requestId, err)
		//Delete the request from memory afterwards
		request.DeployRequests.Delete(requestId)
	}
}
