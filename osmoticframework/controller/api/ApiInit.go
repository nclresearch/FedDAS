package api

import (
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"math"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/api/impl/callback"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/api/impl/request/monitor"
	"osmoticframework/controller/auto"
	"osmoticframework/controller/database"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
	"time"
)

var responseQueue amqp.Queue
var registerQueue amqp.Queue
var alertQueue amqp.Queue

/*
Sets up the response queue and listens to all agent respond.
The response messages are in separate queues to prevent clashing with request messages
This function must be run at last in the main thread. If you need to run something else alongside the listener, please use the goroutine in Controller.go
*/
func Init() {
	var err error
	//Queue and consumer declaration
	responseQueue, err = queue.DeclareControllerQueue("response", true)
	if err != nil {
		log.Fatal.Println("Failed declaring queue response")
		log.Fatal.Panicln(err)
	}
	registerQueue, err = queue.DeclareQueue("register", true)
	if err != nil {
		log.Fatal.Println("Failed declaring queue register")
		log.Fatal.Panicln(err)
	}
	pongQueue, err := queue.DeclareControllerQueue("pong", true)
	if err != nil {
		log.Fatal.Println("Failed declaring queue pong")
		log.Fatal.Panicln(err)
	}
	alertQueue, err = queue.DeclareControllerQueue("alert", true)
	if err != nil {
		log.Fatal.Println("Failed declaring queue alert")
		log.Fatal.Panicln(err)
	}
	customInQueue, err := queue.DeclareControllerQueue("custom-in", true)
	if err != nil {
		log.Fatal.Println("Failed declaring queue custom-in")
		log.Fatal.Panicln(err)
	}

	responseStream, err := queue.NewConsumer(responseQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed creating response consumer")
		log.Fatal.Panicln(err)
	}

	regStream, err := queue.NewConsumer(registerQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed creating register consumer")
		log.Fatal.Panicln(err)
	}

	pongStream, err := queue.NewConsumer(pongQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed creating pong consumer")
		log.Fatal.Panicln(err)
	}

	alertStream, err := queue.NewConsumer(alertQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed creating alert consumer")
		log.Fatal.Panicln(err)
	}

	customInStream, err := queue.NewConsumer(customInQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed creating event consumer")
		log.Fatal.Panicln(err)
	}

	startRoutines(responseStream, regStream, pongStream, alertStream, customInStream)
}

//Go routines for the API
//Channel stream handlers must not jump out of the loop as the controller will stop receiving messages
func startRoutines(responseStream, regStream, pongStream, alertStream, customInStream <-chan amqp.Delivery) {
	//Register listener
	go func() {
		//Kick start the registration thread
		//This must run under a go function so that the controller can stay connected to the channel
		go func() { callback.RegisterThread() }()
		for message := range regStream {
			jsonMsg := deserialize(message.Body)
			//fmt.Println(string(message.Body))
			if jsonMsg != nil {
				//Ignore all messages fired from itself.
				if jsonMsg["direction"] == callback.AGENT {
					continue
				}
				_ = message.Ack(true)
				var currentRequest request.RegisterRequest
				err := json.Unmarshal(message.Body, &currentRequest)
				if err != nil {
					log.Error.Println("Invalid registration info. Ignoring")
					log.Error.Println(err)
					continue
				}
				callback.RegisterQueue <- currentRequest
			}
		}
		if !vars.IsTerminate() {
			log.Fatal.Panicln("Registration stream ended. This is not supposed to happen!")
		}
	}()

	//API response
	go func() {
		for message := range responseStream {
			jsonMsg := deserialize(message.Body)
			//fmt.Println(string(message.Body))
			if jsonMsg != nil {
				api, ok := jsonMsg["api"].(string)
				if !ok {
					//No API specified. Ignore
					continue
				}
				_ = message.Ack(true)
				switch api {
				//This must run under a go function so that the controller can stay connected to the channel
				case "deploy":
					go func() { callback.ParseDeploy(jsonMsg) }()
				case "monitor":
					go func() { callback.ParseMonitor(jsonMsg) }()
				default:
					log.Error.Println("Agent sent unknown response. Ignoring")
				}
			}
		}
		if !vars.IsTerminate() {
			log.Fatal.Panicln("Response stream ended. This is not supposed to happen!")
		}
	}()

	//Keep alive listener
	go func() {
		go func() {
			//Pause for 30 seconds before first alive check
			var seq int64 = 0
			for {
				message, _ := json.Marshal(map[string]interface{}{
					"ping": time.Now().UnixMilli(),
					"seq":  seq,
				})
				seq++
				err := queue.Ch.Publish(
					"",
					"ping",
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        message,
					})
				if err != nil {
					log.Error.Println("Failed sending ping")
					log.Error.Println(err)
				}
				deadAgents := make(map[string]types.Agent)
				vars.Agents.Range(func(agentId, agent interface{}) bool {
					lastAlive := time.Unix(agent.(types.Agent).LastAlive, 0)
					//An agent is considered dead if not connected to controller for more than 30 seconds
					if math.Abs(time.Now().Sub(lastAlive).Seconds()) >= 30 {
						deadAgents[agentId.(string)] = agent.(types.Agent)
					}
					return true
				})
				for agentId, agent := range deadAgents {
					log.Error.Printf("Agent %s has disconnected\n", agentId)
					//Push alert to channel
					alert.AgentDisconnect <- types.OfflineAgent{ID: agentId, Agent: agent}
					//Delete the agent from memory and database
					database.Unregister(agentId)
				}
				time.Sleep(time.Second * 5)
			}
		}()
		for message := range pongStream {
			jsonMsg := deserialize(message.Body)
			if jsonMsg != nil {
				_ = message.Ack(true)
				//Update alive status in another thread
				go callback.ProcessPing(jsonMsg)
			}
		}
		if !vars.IsTerminate() {
			log.Fatal.Panicln("Ping stream ended. This is not supposed to happen!")
		}
	}()

	//Timeout cleanup
	//Find any requests in the map that timed out and removes them
	//We must put this goroutine to sleep for other routines to modify the map of requests
	go func() {
		for {
			time.Sleep(time.Second * 2)
			request.DeployRequests.Range(func(requestId, value interface{}) bool {
				currentRequest := value.(request.ImplRequestTask)
				diff := time.Now().Sub(currentRequest.Time)
				if diff.Seconds() >= currentRequest.Timeout && !currentRequest.Ack {
					log.Error.Println("Deploy request " + requestId.(string) + " timeout")
					callback.CallbackError(requestId.(string), errors.New("timeout"))
					request.DeployRequests.Delete(requestId)
				}
				return false
			})
			request.MonitorRequests.Range(func(requestId, value interface{}) bool {
				currentRequest := value.(request.ImplRequestTask)
				diff := time.Now().Sub(currentRequest.Time)
				if diff.Seconds() >= currentRequest.Timeout && !currentRequest.Ack {
					log.Error.Println("Monitor request " + requestId.(string) + " timeout")
					callback.CallbackError(requestId.(string), errors.New("timeout"))
					request.MonitorRequests.Delete(requestId)
				}
				return false
			})
			if vars.IsTerminate() {
				break
			}
		}
	}()

	//Alert notifications from agents
	go func() {
		for message := range alertStream {
			jsonMsg := deserialize(message.Body)
			if jsonMsg == nil {
				continue
			}
			_ = message.Ack(true)
			alert.ParseAlert(jsonMsg)
		}
		if !vars.IsTerminate() {
			log.Fatal.Panicln("Alert stream ended. This is not supposed to happen!")
		}
	}()

	//Event stream. For trigger based events for the controller
	//There are cases where certain containers needs to be deployed due to other conditions
	//Those containers can contact the controller via this queue
	go func() {
		for message := range customInStream {
			auto.Events <- string(message.Body)
			_ = message.Ack(true)
		}
	}()

	//Kubernetes event watcher
	//Only monitor specific resources
	//If we simply try to get all events in the cluster. Kubernetes will trim events which means we lose information.
	if vars.GetKuberConfigPath() != "" {
		go monitor.DeploymentWatch()
		go monitor.CronjobWatch()
		go monitor.JobWatch()
	} else {
		log.Warn.Println("Monitoring functions on cloud is disabled")
	}

	//Database error handler
	go func() {
		for err := range alert.DatabaseErrors {
			alert.DBErrorHandler(err)
		}
	}()
}

//Deserialize json to a map
func deserialize(body []byte) map[string]interface{} {
	var jsonMsg map[string]interface{}
	err := json.Unmarshal(body, &jsonMsg)
	if err != nil {
		log.Error.Println("Cannot deserialize message. Ignoring")
		log.Error.Println(err)
		return nil
	}
	return jsonMsg
}
