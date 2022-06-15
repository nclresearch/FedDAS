package api

import (
	"encoding/json"
	"github.com/lithammer/shortuuid"
	"github.com/streadway/amqp"
	"net"
	"osmoticframework/agent/constants"
	"osmoticframework/agent/docker"
	"osmoticframework/agent/log"
	"osmoticframework/agent/types"
	"regexp"
	"time"
)

var ch *amqp.Channel
var deployQueue amqp.Queue
var monitorQueue amqp.Queue
var agentId string
var responseQueue amqp.Queue
var alertQueue amqp.Queue

type RegisterMessageDirection string

var (
	CONTROLLER RegisterMessageDirection = "controller"
	AGENT      RegisterMessageDirection = "agent"
)

func Init() {
	//Initialize the connection
	log.Info.Println("Connecting to RabbitMQ server " + constants.GetRabbitAddress())
	server, err := amqp.Dial(constants.GetRabbitAddress())
	if err != nil {
		panic(err)
	}
	defer server.Close()
	ch, err = server.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	//Register itself to the controller
	register()

	//Start queues and consumers
	deployQueue, err = declareExpireQueue("deploy-"+agentId, 30000)
	if err != nil {
		panic(err)
	}
	monitorQueue, err = declareExpireQueue("monitor-"+agentId, 30000)
	if err != nil {
		panic(err)
	}
	pingQueue, err := declareExpireQueue("ping", 30000)
	if err != nil {
		panic(err)
	}

	responseQueue, err = declareControllerQueue("response", true)
	if err != nil {
		panic(err)
	}

	alertQueue, err = declareControllerQueue("alert", true)
	if err != nil {
		panic(err)
	}

	deployStream, err := newConsumer(deployQueue.Name)
	if err != nil {
		panic(err)
	}
	monitorStream, err := newConsumer(monitorQueue.Name)
	if err != nil {
		panic(err)
	}
	pingStream, err := newConsumer(pingQueue.Name)

	/*
		This channel does not serve any data.
		We need to run more than just the API listener, so we need to use goroutines.
		Since the listener is running on a goroutine. The channel is here to prevent the main thread from terminating.
		When a main thread of the application terminates, all running goroutines will terminate as well no matter if they are still operating.
		If you need to run anything else alongside the listener, use a goroutine before running this function.
		The listener must be fired last in the main thread.
	*/
	forever := make(chan bool)

	startRoutine(deployStream, monitorStream, pingStream)

	log.Info.Println("API startup complete. Awaiting instructions")
	<-forever
}

//Registration
func register() {
	log.Info.Println("Registering")
	registerQueue, err := declareQueue("register", true)
	if err != nil {
		log.Fatal.Println("failed declare queue during registration")
		log.Fatal.Panicln(err)
		return
	}
	regStream, err := newConsumer(registerQueue.Name)
	if err != nil {
		log.Fatal.Println("Failed starting consumer during registration")
		log.Fatal.Panicln(err)
		return
	}

	//Construct registration message
	helloId := shortuuid.New()
	hello, err := json.Marshal(map[string]interface{}{
		"requestId":     helloId,
		"direction":     ">>",
		"internalIP":    getInternalIP(constants.GetNetworkInterface()),
		"devSupport":    constants.GetDeviceSupport(),
		"sensorSupport": constants.GetSensorSupport(),
	})
	if err != nil {
		log.Fatal.Println("Failed composing register request")
		log.Fatal.Panicln(err)
		return
	}
	err = ch.Publish(
		"",
		registerQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        hello,
		},
	)
	if err != nil {
		log.Fatal.Println("Failed requesting registration")
		log.Fatal.Panicln(err)
		return
	}

	//Wait for response
	var consumeTag string

	//Check for timeout
	go func() {
		time.Sleep(time.Second * 10)
		if agentId == "" {
			log.Fatal.Panicln("Registration timeout")
		}
	}()

	for message := range regStream {
		var jsonMsg map[string]interface{}
		err := json.Unmarshal(message.Body, &jsonMsg)
		//fmt.Println(string(message.Body))
		if err != nil {
			log.Error.Println("Cannot deserialize registration message. Ignoring")
			log.Error.Println(err)
			continue
		}
		if jsonMsg["requestId"] != helloId || jsonMsg["direction"] == CONTROLLER {
			continue
		}
		//Consume (In RabbitMQ terms, acknowledge) the message and removes it from the queue
		//Otherwise the message will stay at the queue and resend if anyone reconnects.
		_ = message.Ack(true)
		registerFinish := false
		switch jsonMsg["status"] {
		case "success":
			agentId = jsonMsg["agentId"].(string)
			log.Info.Println("Registration successful")
			registerFinish = true
		case "error":
			log.Fatal.Println("Registration rejected")
			log.Fatal.Panicln(jsonMsg["error"].(string))
		}
		consumeTag = message.ConsumerTag
		if registerFinish {
			break
		}
	}
	//Remove the consumer
	_ = ch.Cancel(consumeTag, false)

	log.Info.Println(">> Registered as agent " + agentId)
	return
}

//Starts all goroutines
//In the event of panics, all containers that were still deployed have to be manually stopped.
func startRoutine(deployStream, monitorStream, pingStream <-chan amqp.Delivery) {
	//API listener
	//Deploy API
	go func() {
		for message := range deployStream {
			jsonMsg := deserialize(message.Body)
			//fmt.Println(string(message.Body))
			_ = message.Ack(true)
			if jsonMsg != nil {
				parseDeploy(jsonMsg)
			}
		}
		log.Fatal.Panicln("Deploy stream ended. This should not happen!")
		//If the program somehow reaches here, something has gone wrong with the agent.
	}()
	//Monitoring API
	go func() {
		for message := range monitorStream {
			jsonMsg := deserialize(message.Body)
			_ = message.Ack(true)
			if jsonMsg != nil {
				parseMonitor(jsonMsg)
			}
		}
		log.Fatal.Panicln("Monitor stream ended. This should not happen!")
	}()

	//Heartbeat (ping)
	go func() {
		for message := range pingStream {
			jsonMsg := deserialize(message.Body)
			//Multiple Ack must be false, otherwise messages to other agents will be lost
			_ = message.Ack(false)
			if jsonMsg != nil {
				pongTime := time.Now().UnixMilli()
				latency := pongTime - int64(jsonMsg["ping"].(float64))
				message, _ := json.Marshal(map[string]interface{}{
					"agentId": agentId,
					"pong":    pongTime,
					"seq":     int64(jsonMsg["seq"].(float64)),
					"latency": latency,
				})
				_ = ch.Publish(
					"",
					"pong",
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        message,
					},
				)
			}
		}
		log.Fatal.Panicln("Ping stream ended. This should not happen!")
	}()

	//Automatically deploy monitoring applications
	//Note: The agent will notify the controller it deployed them
	//Since the controller did not request the containers (meaning no request ID), the agent notifies with the request ID "internal"
	go func() {
		log.Info.Println("Deploying monitoring services")
		for _, deployArg := range constants.DefaultContainers {
			log.Info.Println("Self deploying " + deployArg.Image)
			//Start the container
			containerId, err := docker.Run(deployArg, types.AuthInfo{})
			if err != nil {
				replyDeployError("internal", err)
				panic(err)
				return
			}
			//Construct response
			response, _ := json.Marshal(map[string]string{
				"requestId":   "internal",
				"status":      "ok",
				"containerId": containerId,
				"api":         "deploy",
			})
			_ = ch.Publish(
				"",
				responseQueue.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        response,
				},
			)
			log.Info.Println("<< Container " + containerId + " deployed")
		}
		log.Info.Println("Self deploying Prometheus")
		//Start the container
		containerId, err := docker.Run(constants.PrometheusContainer(), types.AuthInfo{})
		if err != nil {
			replyDeployError("internal", err)
			panic(err)
			return
		}
		//Construct response
		response, _ := json.Marshal(map[string]string{
			"requestId":   "internal",
			"status":      "ok",
			"containerId": containerId,
			"api":         "deploy",
		})
		_ = ch.Publish(
			"",
			responseQueue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        response,
			},
		)
		log.Info.Println("<< Container " + containerId + " deployed")
	}()

	//Health checking
	//Checks if the containers are still healthy, and alert the controller when something wrong happens
	go healthCheck()
}

//Deserialize json to a map string interface, where we assert the type of the interface (The value of the map) to any type.
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

//Declares a queue that limits the queue with only one consumer
//When a queue is declared in configuration A, all members who wants to use this queue must also declare with the same config
func declareControllerQueue(queueName string, durable bool) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		queueName,
		durable,
		false,
		false,
		false,
		map[string]interface{}{
			"x-single-active-consumer": true,
		},
	)
	return queue, err
}

//Declares a queue under a name
func declareQueue(queueName string, durable bool) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		queueName,
		durable,
		false,
		false,
		false,
		nil,
	)
	return queue, err
}

//Declares a queue that expires if no messages are received after `expireTime`
func declareExpireQueue(queueName string, expireTime int) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		map[string]interface{}{
			//Expires if there is no consumer connected to the queue for [expireTime] seconds.
			"x-expires": expireTime,
		},
	)
	return queue, err
}

//Starts a new consumer
func newConsumer(queueName string) (<-chan amqp.Delivery, error) {
	consumer, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	return consumer, err
}

//Gets internal IP of an network interface. Required for the controller to determine certain service address
//Only supports IPv4.
func getInternalIP(networkInterface string) string {
	inet, err := net.Interfaces()
	if err != nil {
		log.Fatal.Println("Failed getting network interface")
		log.Fatal.Panicln(err)
	}
	for _, i := range inet {
		if i.Name == networkInterface {
			address, err := i.Addrs()
			if err != nil {
				log.Fatal.Println("Failed getting internal IP from interface " + i.Name)
				log.Fatal.Panicln(err)
			}
			//An interface might provide both IPv4 and IPv6. In this case, we don't need IPv6.
			for _, ip := range address {
				regex := regexp.MustCompile(`^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))/(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
				if regex.MatchString(ip.String()) {
					matches := regex.FindAllStringSubmatch(ip.String(), -1)
					log.Info.Printf("IPv4 address for %s: %s\n", networkInterface, matches[0][1])
					return matches[0][1]
				}
			}
		}
	}
	log.Fatal.Panicln("IPv4 address does not exist for this network interface: " + networkInterface)
	return ""
}

func replyError(requestId string, api string, err error) []byte {
	response, _ := json.Marshal(map[string]string{
		"agentId":   agentId,
		"requestId": requestId,
		"status":    "failed",
		"api":       api,
		"error":     err.Error(),
	})
	return response
}

func replyAck(requestId, apiName string) {
	ack, _ := json.Marshal(map[string]interface{}{
		"requestId": requestId,
		"status":    "ack",
		"api":       apiName,
	})
	err := ch.Publish(
		"",
		responseQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        ack,
		},
	)
	if err != nil {
		log.Error.Println("Failed pushing ack")
		log.Error.Println(err)
	}
}
