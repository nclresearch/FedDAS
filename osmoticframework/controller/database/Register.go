package database

import (
	"database/sql"
	"github.com/lithammer/shortuuid"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/log"
	"osmoticframework/controller/queue"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
	"time"
)

//Generate a new agent ID with duplicate checks
func GenerateAgentID() string {
	agentId := shortuuid.New()
	//Check if agentId exists in database
	query := "SELECT * FROM registry WHERE AgentId = ?"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{agentId},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return ""
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Error.Println("Error preparing statement: " + err.Error())
		return ""
	}
	//goland:noinspection GoUnhandledErrorResult
	defer stmt.Close()
	var agentIdExists bool
	err = stmt.QueryRow(agentId).Scan(&agentIdExists)
	if err != nil && err != sql.ErrNoRows {
		log.Error.Println("Error querying database: " + err.Error())
		return ""
	}
	if agentIdExists {
		return GenerateAgentID()
	}
	return agentId
}

//Registers a new agent to the database
//All agents information are stored in measurement "registry" in database "agents"
func Register(agentId, internalIP string, devSupport, sensorSupport []string) error {
	query := "INSERT INTO registry (AgentId, InternalIP) VALUES (?, ?)"
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     query,
			QueryArgs: []string{},
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return err
	}
	defer db.Close()
	//Write agent registry to database
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	timestamp := time.Now()
	_, err = stmt.Exec(agentId, internalIP, timestamp.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}

	for _, dev := range devSupport {
		query := "INSERT INTO devSupport (AgentId, Device) VALUES (? ,?)"
		stmt, err = db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId, dev},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		_, err = stmt.Exec(agentId, dev)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId, dev},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
		}
	}

	for _, sensor := range sensorSupport {
		query := "INSERT INTO sensorSupport (agentid, sensor) VALUES (?, ?)"
		stmt, err = db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId, sensor},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
			continue
		}
		_, err = stmt.Exec(db, sensor)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId, sensor},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
		}
	}

	//Record the agent ID in memory. The controller uses the copy from memory to deploy containers.
	//The database is used when the controller fails and needs to recover.
	//Store the registered agent to memory as well
	newAgent := types.Agent{
		LastAlive:     timestamp.Unix(),
		DeviceSupport: devSupport,
		SensorSupport: sensorSupport,
		InternalIP:    internalIP,
		Containers:    make([]string, 0),
		PingSeq:       0,
	}
	vars.Agents.Store(agentId, newAgent)

	queue.SetupAgentQueue(agentId)
	return nil
}

//Removes agent from the database
func Unregister(agentId string) {
	queries := [4]string{
		"DELETE FROM agents.containers WHERE AgentId = ?",
		"DELETE FROM agents.devSupport WHERE AgentId = ?",
		"DELETE FROM agents.sensorSupport WHERE AgentId = ?",
		"DELETE FROM agents.registry WHERE AgentId = ?",
	}
	db, err := sql.Open("mysql", vars.GetDatabaseAddress())
	if err != nil {
		alert.DatabaseErrors <- types.DatabaseErrorReport{
			Query:     "",
			QueryArgs: nil,
			Error:     err,
		}
		log.Error.Println("Error occurred during writing to database")
		log.Error.Println(err)
		return
	}
	defer db.Close()
	for _, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			return
		}
		_, err = stmt.Exec(agentId)
		if err != nil {
			alert.DatabaseErrors <- types.DatabaseErrorReport{
				Query:     query,
				QueryArgs: []string{agentId},
				Error:     err,
			}
			log.Error.Println("Error occurred during writing to database")
			log.Error.Println(err)
		}
	}
	vars.Agents.Delete(agentId)
}
