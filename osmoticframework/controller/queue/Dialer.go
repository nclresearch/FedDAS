package queue

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/streadway/amqp"
	"io/ioutil"
	"osmoticframework/controller/log"
	"osmoticframework/controller/vars"
	"path/filepath"
	"strings"
)

var Server *amqp.Connection
var Ch *amqp.Channel
var deployQueue []amqp.Queue

/*
Connects to the RabbitMQ server. This function must be called first before sending any requests.
This sets up the channel and queue needed for sending request to agents.
It does NOT setup the queue to receive response messages. See ApiInit.go.
*/
func Init() {
	log.Info.Println("Connecting to server " + vars.GetRabbitAddress())
	if strings.HasPrefix(vars.GetRabbitAddress(), "amqp://") {
		dial()
	} else if strings.HasPrefix(vars.GetRabbitAddress(), "amqps://") {
		dialTls()
	} else {
		log.Fatal.Fatalln("Invalid RabbitMQ address. Address must start with amqp:// or amqps://")
	}
}

func dial() {
	var err error
	Server, err = amqp.Dial(vars.GetRabbitAddress())
	if err != nil {
		log.Fatal.Println("Failed to dial RabbitMQ")
		log.Fatal.Panicln(err)
	}
	Ch, err = Server.Channel()
	if err != nil {
		log.Fatal.Println("Failed to dial RabbitMQ")
		log.Fatal.Panicln(err)
	}
}

func dialTls() {
	tlsConf := new(tls.Config)
	//RootCA
	tlsConf.RootCAs = x509.NewCertPool()
	cacert := filepath.Join(vars.GetCredDirectory(), "ca_certificate.pem")
	ca, err := ioutil.ReadFile(cacert)
	if err != nil {
		log.Fatal.Println("Failed to load CA certificate")
		log.Fatal.Panicln(err)
	}
	tlsConf.RootCAs.AppendCertsFromPEM(ca)
	//Load key pair
	clientCert := filepath.Join(vars.GetCredDirectory(), "client_certificate.pem")
	clientKey := filepath.Join(vars.GetCredDirectory(), "client_key.pem")
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatal.Println("Failed to load client certificate")
		log.Fatal.Panicln(err)
	}
	tlsConf.Certificates = append(tlsConf.Certificates, cert)
	Server, err := amqp.DialTLS(vars.GetRabbitAddress(), tlsConf)
	if err != nil {
		log.Fatal.Println("Failed to dial RabbitMQ")
		log.Fatal.Panicln(err)
	}
	log.Info.Println("Connected to RabbitMQ server")
	Ch, err = Server.Channel()
	if err != nil {
		log.Fatal.Println("Failed to dial RabbitMQ")
		log.Fatal.Panicln(err)
	}
	log.Info.Println("Channel created")
}

//Declares a queue on RabbitMQ, limits to only one consumer
func DeclareControllerQueue(queueName string, durable bool) (amqp.Queue, error) {
	queue, err := Ch.QueueDeclare(
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

//Declares a queue on RabbitMQ
func DeclareQueue(queueName string, durable bool) (amqp.Queue, error) {
	queue, err := Ch.QueueDeclare(
		queueName,
		durable,
		false,
		false,
		false,
		nil,
	)
	return queue, err
}

//Declares a queue in RabbitMQ, which expires if no messages are sent within expireTime
func DeclareExpireQueue(queueName string, expireTime int) (amqp.Queue, error) {
	queue, err := Ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		map[string]interface{}{
			//Expires when there are no consumers connected
			"x-expires": expireTime,
		},
	)
	return queue, err
}

//Declares a consumer using the queue
func NewConsumer(queueName string) (<-chan amqp.Delivery, error) {
	consumer, err := Ch.Consume(
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

//Sets up the queues for the agent.
func SetupAgentQueue(agentId string) {
	deploy, _ := DeclareExpireQueue("deploy-"+agentId, 30000)
	deployQueue = append(deployQueue, deploy)
}
