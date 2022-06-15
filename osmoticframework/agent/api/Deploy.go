package api

import (
	"encoding/json"
	"errors"
	api "github.com/fsouza/go-dockerclient"
	"osmoticframework/agent/alert"
	"osmoticframework/agent/docker"
	"osmoticframework/agent/log"
	"osmoticframework/agent/types"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

//Deploy API endpoint
//This is the request handler of the deploy API. For processing requests, go to docker/Deployer.go

//Creates and starts a container
//Pulls the image if it does not exist
func RunEP(requestId string, args map[string]interface{}) []byte {
	if args["deployArgs"] == nil {
		replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	//Decode the generic interface map to DeployArgs struct
	var deployArgs types.DeployArgs
	err := mapstructure.Decode(args["deployArgs"], &deployArgs)
	if err != nil {
		return replyDeployError(requestId, err)
	}
	var authInfo types.AuthInfo
	if args["authInfo"] != nil {
		err = mapstructure.Decode(args["authInfo"], &authInfo)
		if err != nil {
			return replyDeployError(requestId, err)
		}
	}

	//Start the container
	containerId, err := docker.Run(deployArgs, authInfo)
	if err != nil {
		return replyDeployError(requestId, err)
	}

	//Construct response
	response, _ := json.Marshal(map[string]string{
		"requestId":   requestId,
		"status":      "ok",
		"containerId": containerId,
		"api":         "deploy",
	})
	log.Info.Println("<< Container " + containerId + " deployed")
	return response
}

//Stops a container
func StopEP(requestId string, args map[string]interface{}) []byte {
	containerId, ok := args["containerId"].(string)
	if !ok {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	err := docker.Stop(containerId)
	if err != nil {
		return replyDeployError(requestId, err)
	}

	log.Info.Println("<< Stopped container " + containerId)
	return replyDeployOk(requestId)
}

//Deletes a container
//You cannot delete a container when it's running. Stop it first.
func DeleteEP(requestId string, args map[string]interface{}) []byte {
	containerId, ok := args["containerId"].(string)
	if !ok {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	deleteImage, ok := args["deleteImage"].(bool)
	if !ok {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	err := docker.Delete(containerId, deleteImage)
	if err != nil {
		return replyDeployError(requestId, err)
	}

	log.Info.Println("<< Removed container " + containerId)
	return replyDeployOk(requestId)
}

//Update container
func UpdateEP(requestId string, args map[string]interface{}) []byte {
	if args["deployArgs"] == nil {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	containerId, ok := args["containerId"].(string)
	if !ok {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	log.Info.Println("Updating container " + containerId)
	//Decode the generic interface map to DeployArgs struct
	var deployArgs types.DeployArgs
	err := mapstructure.Decode(args["deployArgs"], &deployArgs)
	if err != nil {
		return replyDeployError(requestId, err)
	}
	var authInfo types.AuthInfo
	if args["authInfo"] != nil {
		err = mapstructure.Decode(args["authInfo"], &authInfo)
		if err != nil {
			return replyDeployError(requestId, err)
		}
	}
	newContainerId, err := docker.Update(containerId, deployArgs, authInfo)
	if err != nil {
		return replyDeployError(requestId, err)
	}

	//Construct response
	response, _ := json.Marshal(map[string]string{
		"requestId":   requestId,
		"status":      "ok",
		"containerId": newContainerId,
		"api":         "deploy",
	})

	log.Info.Println("<< Container " + containerId + " is now updated to " + newContainerId)
	return response
}

//Lists and inspects all containers
func ListEP(requestId string) []byte {
	containers, err := docker.ListDetailed()
	if err != nil {
		return replyDeployError(requestId, err)
	}
	response, _ := json.Marshal(map[string]interface{}{
		"requestId":  requestId,
		"status":     "ok",
		"api":        "deploy",
		"containers": containers,
	})

	log.Info.Println("<< Listing containers")
	return response
}

//Inspect container
func InspectEP(requestId string, args map[string]interface{}) []byte {
	containerId, ok := args["containerId"].(string)
	if !ok {
		return replyDeployError(requestId, errors.New("cannot parse arguments"))
	}
	container, err := docker.Inspect(containerId)
	if err != nil {
		return replyDeployError(requestId, err)
	}
	response, _ := json.Marshal(map[string]interface{}{
		"requestId": requestId,
		"status":    "ok",
		"api":       "deploy",
		"container": *container,
	})

	log.Info.Printf("<< Inspecting container %s\n", containerId)
	return response
}

func healthCheck() {
	//Filter down events to only specific events about containers
	options := api.EventsOptions{Filters: map[string][]string{
		"type": {"container"},
		//create = creation of the container
		//stop = container stop
		//start = container start
		//die = container exit (or crash)
		//destroy = container removal
		"event": {"create", "stop", "start", "die", "destroy"},
	}}
	for {
		//Create listener
		listener := make(chan *api.APIEvents)
		err := docker.RegisterListener(options, listener)
		if err != nil {
			log.Error.Println("Failed starting Docker event listener. Retrying in 5 seconds")
			time.Sleep(time.Second * 5)
			continue
		}
		for event := range listener {
			switch event.Action {
			case "create":
				log.Info.Printf("Container %s created\n", event.Actor.ID)
			case "stop":
				log.Info.Printf("Container %s stopped\n", event.Actor.ID)
			case "start":
				log.Info.Printf("Container %s started\n", event.Actor.ID)
			case "die":
				exitCode, _ := strconv.ParseInt(event.Actor.Attributes["exitCode"], 10, 64)
				crash := types.CrashReport{
					AgentId:  agentId,
					ID:       event.Actor.ID,
					Name:     event.Actor.Attributes["name"],
					Image:    event.Actor.Attributes["image"],
					ExitCode: int(exitCode),
				}
				if exitCode == 0 {
					log.Info.Printf("Container %s has exited with code 0\n", event.Actor.ID)
					crash.Status = "exited"
				} else {
					log.Warn.Printf("Container %s has crashed with exit code %d\n", event.Actor.ID, exitCode)
					crash.Status = "error"
				}
				response, _ := json.Marshal(map[string]interface{}{
					"type":     alert.AlertContainerCrash,
					"contents": crash,
				})
				//Publish message to controller
				err = ch.Publish(
					"",
					alertQueue.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        response,
					},
				)
				if err != nil {
					log.Error.Println("Failed pushing container alert")
					log.Error.Println(err)
				}
			case "destroy":
				log.Info.Printf("Container %s removed\n", event.Actor.ID)
			}
		}
		log.Warn.Printf("Docker has hung up. Restarting event stream")
		//Deregister the listener and set up a new one
		err = docker.DeregisterListener(listener)
		if err != nil {
			log.Error.Println("Failed to remove listener. This may cause memory leak")
		}
	}
}

//Processes deploy commands only
func parseDeploy(jsonMsg map[string]interface{}) {
	requestId, ok := jsonMsg["requestId"].(string)
	if !ok {
		//We can't figure what the request is. Ignore completely
		return
	}
	args, ok := jsonMsg["args"].(map[string]interface{})
	if !ok {
		//fmt.Println(jsonMsg)
		replyDeployError(requestId, errors.New("deploy - cannot parse request arguments"))
		return
	}
	replyAck(requestId, "deploy")
	//As deployments can take a long time, these operations are done in a separate go function so that the agent can process other requests in parallel
	//Also the RabbitMQ server will disconnect the agent if the client does not process the message on time
	go func() {
		var response []byte
		switch jsonMsg["command"] {
		case "run":
			response = RunEP(requestId, args)
		case "stop":
			response = StopEP(requestId, args)
		case "update":
			response = UpdateEP(requestId, args)
		case "delete":
			response = DeleteEP(requestId, args)
		case "list":
			response = ListEP(requestId)
		case "inspect":
			response = InspectEP(requestId, args)
		default:
			response = replyDeployError(requestId, errors.New("unknown command"))
		}
		err := ch.Publish(
			"",
			responseQueue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        response,
			},
		)
		if err != nil {
			log.Error.Println("Failed pushing response")
			log.Error.Println(err)
		}
	}()
}

func replyDeployError(requestId string, err error) []byte {
	return replyError(requestId, "deploy", err)
}

func replyDeployOk(requestId string) []byte {
	response, _ := json.Marshal(map[string]string{
		"requestId": requestId,
		"status":    "ok",
		"api":       "deploy",
	})
	return response
}
