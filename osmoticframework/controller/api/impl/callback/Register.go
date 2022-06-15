package callback

import (
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/auto"
	"osmoticframework/controller/database"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
)

//Registration queue. All unprocessed registration requests are stored here.
var RegisterQueue = make(chan request.RegisterRequest)

type RegisterMessageDirection string

var (
	CONTROLLER RegisterMessageDirection = "controller"
	AGENT      RegisterMessageDirection = "agent"
)

//Registration flow goes like the following:
//Receiving request -> record on database -> Check if agent's network already has monitor API enabled ->
//If not enabled -> record on database -> deploy prometheus on agent -> send agent ID to agent
//If enabled -> send agent ID to agent
//Due to the need to deploy Prometheus on ONLY ONE agent for each LAN. Registration must be done via a queue (in this case the requests are built up inside a channel)
//Otherwise there will be concurrency problems where there are multiple agents deploying Prometheus, wasting resources.

func RegisterThread() {
	//If there are no registration request in the queue, the thread simply goes to sleep
	for regRequest := range RegisterQueue {
		log.Info.Printf("(reg: %s) >> Registering agent\n", regRequest.ID)
		//Generate ID
		agentId := database.GenerateAgentID()
		if agentId == "" {
			log.Error.Printf("(reg: %s) >> Failed to generate agent ID\n", regRequest.ID)
			rejectRegistration(regRequest.ID, errors.New("failed to generate agent ID"))
			continue
		}
		//Register to database
		err := database.Register(agentId, regRequest.InternalIP, regRequest.DeviceSupport, regRequest.SensorSupport)
		if err != nil {
			rejectRegistration(regRequest.ID, err)
			continue
		}
		//Send response back to agent
		response, _ := json.Marshal(map[string]string{
			"requestId": regRequest.ID,
			"direction": string(CONTROLLER),
			"agentId":   agentId,
			"status":    "success",
		})
		err = queue.Ch.Publish(
			"",
			"register",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        response,
			},
		)
		if err != nil {
			log.Error.Println("Failed sending registration info")
			log.Error.Println(err)
			database.Unregister(agentId)
			continue
		}
		log.Info.Printf("%s (reg: %s) << Registered agent\n", agentId, regRequest.ID)
		auto.AgentJoin <- agentId
	}
	log.Fatal.Fatalln("Registration thread ended. This should not happen!")
}

func rejectRegistration(requestId string, err error) {
	log.Error.Printf("(reg: %s) << Registration failure", requestId)
	response, _ := json.Marshal(map[string]string{
		"requestId": requestId,
		"status":    "error",
		"error":     err.Error(),
	})
	_ = queue.Ch.Publish(
		"",
		"register",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        response,
		},
	)
}
