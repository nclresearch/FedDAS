package callback

import (
	"osmoticframework/controller/alert"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
	"time"
)

func ProcessPing(message map[string]interface{}) {
	agentId, ok := message["agentId"].(string)
	if !ok {
		// Broken ping message?
		return
	}
	agent, ok := vars.Agents.Load(agentId)
	if !ok {
		// Disconnected agent still sending ping?
		return
	}
	seq, ok := message["seq"].(int64)
	if !ok {
		// Broken ping message?
		return
	}
	// Newer ping arrived before older one. Ignore.
	if seq < agent.(types.Agent).PingSeq {
		return
	}
	// Pong contains the timestamp when the agent received the ping. Currently unused
	_, ok = message["pong"].(int64)
	if !ok {
		return
	}
	latency, ok := message["latency"].(int64)
	if !ok {
		// Broken ping message?
		return
	}
	// Latency longer than 3 seconds
	if latency > 3000 {
		// Possible long latency
		alert.PerformanceIssues <- types.PerformanceReport{
			AgentId:   agentId,
			CloudSide: false,
			AlertType: types.PerformanceAlertTypeNetworkLatency,
		}
		log.Warn.Printf("Agent %s has a high latency of %d ms\n", agentId, latency)
	}
	//Update the last alive in memory
	//We can't change a single element of a struct inside a map. The whole struct needs to be rewritten to the map
	vars.Agents.Store(agentId, types.Agent{
		InternalIP:    agent.(types.Agent).InternalIP,
		DeviceSupport: agent.(types.Agent).DeviceSupport,
		SensorSupport: agent.(types.Agent).SensorSupport,
		Containers:    agent.(types.Agent).Containers,
		LastAlive:     time.Now().Unix(),
		PingSeq:       seq + 1,
	})
}
