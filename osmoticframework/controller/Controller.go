package main

import (
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"osmoticframework/controller/api"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/auto"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"osmoticframework/controller/recovery"
	"osmoticframework/controller/types"
	_ "osmoticframework/controller/util"
	"osmoticframework/controller/vars"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var sigterm = make(chan os.Signal)

func main() {
	log.Info.Println("Credential directory: " + vars.GetCredDirectory())
	jsonFile, err := os.Open("properties.json")
	if err != nil {
		log.Fatal.Println("Properties file not found. Please refer to documentation")
		log.Fatal.Fatalln("https://github.com/lukewen427/osmoticframework/wiki/Agent-and-controller", err)
	}
	log.Info.Println("Loading config file")
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	vars.LoadConfig(bytes)
	_ = jsonFile.Close()
	log.Info.Println("Cloud cluster IP: " + vars.GetListeningIP())
	log.Info.Println("RabbitMQ address: " + vars.GetRabbitAddress())
	log.Info.Println("MySQL address: " + vars.GetDatabaseAddress())
	log.Info.Println("Cloud Prometheus address: " + vars.GetPrometheusAddress())
	if vars.IsProfilerEnable() {
		log.Info.Println("Serving controller profiler at localhost:" + strconv.Itoa(vars.GetProfilerPort()))
		log.Warn.Println("Turning on profiling can affect performance. Consider turning it off in production")
		go func() {
			//Limit the serving scope to localhost only
			_ = http.ListenAndServe("localhost:"+strconv.Itoa(vars.GetProfilerPort()), nil)
		}()
	}
	//This is a simple TCP port check. It doesn't account for any handshake issues later on.
	//This is to prevent the controller from immediately crashing because RabbitMQ or MySQL is still starting up
	log.Info.Println("Waiting for RabbitMQ (15s)...")
	rabbitChan := make(chan bool)
	go func() {
		timeout := 15
		for i := 0; i < timeout; i++ {
			address := net.JoinHostPort(vars.GetRabbitHost(), vars.GetRabbitPort())
			conn, err := net.Dial("tcp", address)
			if err == nil {
				_ = conn.Close()
				rabbitChan <- true
				return
			}
			time.Sleep(time.Second)
		}
		rabbitChan <- false
	}()
	log.Info.Println("Waiting for MySQL (15s)...")
	mysqlChan := make(chan bool)
	go func() {
		timeout := 15
		for i := 0; i < timeout; i++ {
			conn, err := net.Dial("tcp", net.JoinHostPort(vars.GetMysqlHost(), vars.GetMysqlPort()))
			if err == nil {
				_ = conn.Close()
				mysqlChan <- true
				return
			}
			time.Sleep(time.Second)
		}
		mysqlChan <- false
	}()
	if !<-rabbitChan {
		log.Fatal.Panicln("RabbitMQ: Host unreachable/Timeout")
	}
	if !<-mysqlChan {
		log.Fatal.Panicln("MySQL: Host unreachable/Timeout")
	}
	//Initialize connections to servers
	queue.Init()
	//Start recovery
	recovery.Recover()
	//Start auto deploy logic
	go auto.AutoMain()
	//Initialize the API
	api.Init()

	//Wait for SIGTERM (Ctrl+C). And start the teardown procedure
	log.Info.Println("Listener startup complete. Listening to response")
	vars.SetApiReady()
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)
	<-sigterm
	log.Info.Println("SIGTERM received. Stopping controller")
	log.Info.Println("Shutting down services...(Press Ctrl+C again to skip)")
	go teardown()
	_, open := <-sigterm
	if open {
		log.Warn.Println("Skipped shutting down services. You will need to clean up the deployed services manually")
	}
	disconnect()
	log.Info.Println("Osmotic controller shutdown. Author: Ringo Sham")
}

func teardown() {
	//3 second window before teardown
	time.Sleep(time.Second * 3)
	vars.Deployments.Range(func(_name interface{}, _ interface{}) bool {
		name := _name.(string)
		err := request.KDeleteDeployment(name)
		if err != nil {
			log.Error.Printf("Failed to delete deployment %s\n", name)
			log.Error.Println(err)
		}
		return true
	})
	vars.Services.Range(func(_name interface{}, _ interface{}) bool {
		name := _name.(string)
		err := request.KDeleteService(name)
		if err != nil {
			log.Error.Printf("Failed to delete services %s\n", name)
			log.Error.Println(err)
		}
		return true
	})
	vars.Jobs.Range(func(_name interface{}, _ interface{}) bool {
		name := _name.(string)
		err := request.KDeleteJob(name)
		if err != nil {
			log.Error.Printf("Failed to delete job %s\n", name)
			log.Error.Println(err)
		}
		return true
	})
	vars.CronJobs.Range(func(_name interface{}, _ interface{}) bool {
		name := _name.(string)
		err := request.KDeleteCronJob(name)
		if err != nil {
			log.Error.Printf("Failed to delete cronjob %s\n", name)
			log.Error.Println(err)
		}
		return true
	})
	log.Info.Println("Cloud service shut down complete")
	log.Info.Println("Note: Pods are still terminating and might need more time before they completely stops")
	//This maps the agent ID to a list of container IDs that this agent has deployed
	agents := make(map[string][]string)
	//We do not want to do stop containers inside this Range function as it will block other threads writing to it
	vars.Agents.Range(func(_agentId interface{}, _agent interface{}) bool {
		agentId := _agentId.(string)
		agent := _agent.(types.Agent)
		agents[agentId] = agent.Containers
		return true
	})
	//Delete containers from all agents
	wg := new(sync.WaitGroup)
	count := 0
	for agentId, containers := range agents {
		for _, containerId := range containers {
			go func(wg *sync.WaitGroup, agentId, containerId string) {
				defer wg.Done()
				task := request.DeleteRequest(agentId, containerId, false, 30)
				result := <-task.Result
				if result.ResultType == request.Error {
					log.Error.Printf("Failed to delete container %s from agent %s\n", containerId, agentId)
					log.Error.Println(result.Content.(error))
				}
			}(wg, agentId, containerId)
			count++
		}
	}
	//Wait groups will wait until all of the requests are complete, regardless of success
	wg.Add(count)
	wg.Wait()
	log.Info.Println("Agent service shut down complete")
	close(sigterm)
}

func disconnect() {
	log.Info.Println("Shutting down APIs")
	vars.SetTerminate()
	//Gracefully disconnect
	//Disconnecting from the servers should close down all of the RabbitMQ streams
	_ = queue.Ch.Close()
	_ = queue.Server.Close()
}
