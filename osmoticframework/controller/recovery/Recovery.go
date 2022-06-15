package recovery

import (
	"database/sql"
	"github.com/mitchellh/mapstructure"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
	"strconv"
	"strings"
	"time"
)

//Recover from database, if the controller has crashed
//If the system is started from a clean state, this function will not do anything.
//The controller should query all recovered agents or the database on what containers are deployed.
func Recover() {
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     "",
			QueryArgs: []string{},
			Error:     err,
		}
		log.Fatal.Println("Error occurred during writing to database")
		log.Fatal.Panicln(err)
		return
	}
	defer db.Close()
	//Recover list of agents from database and write back to memory
	registry, err := db.Query("SELECT * FROM registry")
	if err != nil {
		//Failing to recover would mean a lot of agents operating without controller input. Thus panic.
		log.Fatal.Println("Recovery failed")
		log.Fatal.Panicln(err)
	}
	log.Info.Println("Checking database")
	recoverCount := 0
	for registry.Next() {
		//Get all containers
		var agentId string
		var timestamp time.Time
		var internalIP string
		_ = registry.Scan(&agentId, &timestamp, &internalIP)
		query := "SELECT containers.ContainerId FROM containers WHERE AgentId = ?"
		containerStmt, err := db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		containerQuery, err := containerStmt.Query(agentId)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		var containers []string
		for containerQuery.Next() {
			var container string
			_ = containerQuery.Scan(&container)
			containers = append(containers, container)
		}

		//Get agent's device support
		query = "SELECT devSupport.Device FROM devSupport WHERE AgentId = ?"
		devStmt, err := db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		devQuery, err := devStmt.Query(agentId)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		var devSupport []string
		for devQuery.Next() {
			var device string
			_ = devQuery.Scan(&device)
			devSupport = append(devSupport, device)
		}

		//Get agent's sensor support
		query = "SELECT sensorSupport.Sensor FROM sensorSupport WHERE AgentId = ?"
		sensorStmt, err := db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		sensorQuery, err := sensorStmt.Query(agentId)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		var sensorSupport []string
		for sensorQuery.Next() {
			var sensor string
			_ = sensorQuery.Scan(&sensor)
			sensorSupport = append(sensorSupport, sensor)
		}

		vars.Agents.Store(agentId, types.Agent{
			Containers:    containers,
			LastAlive:     timestamp.UnixNano(),
			InternalIP:    internalIP,
			DeviceSupport: devSupport,
			SensorSupport: sensorSupport,
		})
		recoverCount++
	}

	log.Info.Println("Number of agents registered from recovery: " + strconv.Itoa(recoverCount))

	//End of recovery if there are no agents registered in database
	if recoverCount == 0 {
		return
	}

	//Setup RabbitMQ queues for every agent
	agents := make(map[string][]string, 0)
	vars.Agents.Range(func(_agentId, _agent interface{}) bool {
		agentId := _agentId.(string)
		agent := _agent.(types.Agent)
		queue.SetupAgentQueue(agentId)
		agents[agentId] = agent.Containers
		return true
	})
	var coreImages = [...]string{"osmotic_agent", "prom/prometheus", "prom/node-exporter", "google/cadvisor"}

	//vars.Agents should now contain everything about all registered agents, including the containers deployed (container ID only)
	//The controller needs to inspect these container IDs to get more context later on
	//In the case of agents disconnected while the controller is offline, this will be dealt with when the controller is fully started up
	//The lack of heartbeats will alert the controller to handle the missing agents

	//Delete any containers that doesn't exist in database after the API is ready
	//So that we can free some unused resources
	go func() {
		for !vars.GetApiReady() {
		}
		for aId, aCon := range agents {
			agentId := aId
			localContainers := aCon
			go func(agentId string, localContainers []string) {
				task := request.ListRequest(agentId, 10)
				result := <-task.Result
				if result.ResultType == request.Error {
					log.Error.Printf("!! Failed cleaning agent %s\n", agentId)
					log.Error.Println(result.Content.(error))
					//Ignore if errored
					return
				}
				var containers []types.Container
				err := mapstructure.Decode(result.Content, &containers)
				if err != nil {
					//That shouldn't happen
					log.Fatal.Println("failed deserializing containers")
					log.Fatal.Panicln(err)
				}
				containerToDel := make([]string, 0)
				for _, container := range containers {
					exist := false
					for _, localContainer := range localContainers {
						//Prevent removing critical containers (e.g Prometheus, node exporter, cAdvisor and the agent itself)
						for _, coreImage := range coreImages {
							if strings.HasPrefix(container.Image, coreImage) {
								break
							}
						}
						if container.ID == localContainer {
							exist = true
							break
						}
					}
					if !exist {
						containerToDel = append(containerToDel, container.ID)
					}
				}
				for _, container := range containerToDel {
					request.DeleteRequest(agentId, container, false, 10)
					//Ignore response
				}
			}(agentId, localContainers)
		}
	}()
}
