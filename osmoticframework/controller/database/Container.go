package database

import (
	"database/sql"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
)

/*
	Data entries for adding/removing/etc. containers
	All container information are written to the measurement "containers" in database "agents"
*/

type DBContainer struct {
	ContainerID string
	AgentID     string
	Status      string
}

func AddContainer(agentId, containerId string) {
	query := "INSERT INTO containers (ContainerId, AgentId, Status) VALUES (?, ?, 'running')"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during connecting to database")
		log.Error.Println(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	_, err = stmt.Query(containerId, agentId)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	//Update the agent in memory
	agentInterface, _ := vars.Agents.Load(agentId)
	agent := agentInterface.(types.Agent)
	//Replace the entire agent struct. Since sync.Map does not allow changing variables within a struct
	containers := agent.Containers
	containers = append(containers, containerId)
	vars.Agents.Store(agentId, types.Agent{
		InternalIP:    agent.InternalIP,
		DeviceSupport: agent.DeviceSupport,
		SensorSupport: agent.SensorSupport,
		Containers:    containers,
		LastAlive:     agent.LastAlive,
		PingSeq:       agent.PingSeq,
	})
}

//InfluxDB does not support updating entries. We must delete the row and insert again.
func StopContainer(agentId, containerId string) {
	query := "UPDATE containers SET containers.Status = 'stopped' WHERE AgentId = ? AND ContainerId = ?"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during connecting to database")
		log.Error.Println(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	_, err = stmt.Query(agentId, containerId)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
	}
}

//Remove the row completely.
func RemoveContainer(agentId, containerId string) {
	query := "DELETE FROM containers WHERE AgentId = ? AND ContainerId = ?"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during connecting to database")
		log.Error.Println(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	_, err = stmt.Query(agentId, containerId)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
	}
	//Update the agent in memory
	agentInterface, _ := vars.Agents.Load(agentId)
	agent := agentInterface.(types.Agent)
	//Replace the entire agent struct. Since sync.Map does not allow changing variables within a struct
	containers := agent.Containers
	//Find the array index of where the container ID is stored
	index := 0
	for _, v := range containers {
		if v != containerId {
			containers[index] = v
			index++
		}
	}
	vars.Agents.Store(agentId, types.Agent{
		InternalIP:    agent.InternalIP,
		DeviceSupport: agent.DeviceSupport,
		SensorSupport: agent.SensorSupport,
		//Slice out the removed container ID
		Containers: containers[:index],
		LastAlive:  agent.LastAlive,
		PingSeq:    agent.PingSeq,
	})
}

func UpdateContainer(agentId, oldContainer, containerId string) {
	query := "UPDATE containers SET ContainerId = ? WHERE AgentId = ? AND ContainerId = ?"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId, oldContainer},
			Error:     err,
		}
		log.Error.Println("Error occurred during connecting to database")
		log.Error.Println(err)
		return
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId, oldContainer},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	_, err = stmt.Query(containerId, agentId, oldContainer)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{containerId, agentId, oldContainer},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	//Update the agent in memory
	agentInterface, _ := vars.Agents.Load(agentId)
	agent := agentInterface.(types.Agent)
	//Replace the entire agent struct. Since sync.Map does not allow changing variables within a struct
	containers := agent.Containers
	//Find the array index of where the container ID is stored
	index := 0
	for _, v := range containers {
		if v != oldContainer {
			containers[index] = v
			index++
		}
	}
	//Slice out the removed container ID
	containers = containers[:index]
	//Add the new one back
	containers = append(containers, containerId)
	vars.Agents.Store(agentId, types.Agent{
		InternalIP:    agent.InternalIP,
		DeviceSupport: agent.DeviceSupport,
		SensorSupport: agent.SensorSupport,
		Containers:    containers,
		LastAlive:     agent.LastAlive,
		PingSeq:       agent.PingSeq,
	})
}

//List all containers registered to the database
func ListContainers() []DBContainer {
	query := "SELECT * FROM containers"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{},
			Error:     err,
		}
		log.Error.Println("Error occurred during connecting to database")
		log.Error.Println(err)
		return nil
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{},
			Error:     err,
		}
		log.Error.Println("Error occurred query to database")
		log.Error.Println(err)
		return nil
	}
	rows, err := stmt.Query()
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{},
			Error:     err,
		}
		log.Error.Println("Error occurred query to database")
		log.Error.Println(err)
		return nil
	}
	defer rows.Close()
	containers := make([]DBContainer, 0)
	for rows.Next() {
		var AgentID, ContainerID, Status string
		err := rows.Scan(&ContainerID, &AgentID, &Status)
		if err != nil {
			log.Error.Println("Error occurred reading from database")
			log.Error.Println(err)
			return nil
		}
		containers = append(containers, DBContainer{
			AgentID:     AgentID,
			ContainerID: ContainerID,
			Status:      Status,
		})
	}
	return containers
}
