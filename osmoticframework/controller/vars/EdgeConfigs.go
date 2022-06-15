package vars

import (
	"encoding/base64"
	"encoding/json"
	"osmoticframework/controller/types"
)

//Static variables to perform deploying operations on edge devices

//Generate target file using a list of agent IDs. Encoded in base64
func GenerateTargetJson(agentIds map[string]types.Agent) string {
	target := make([]TargetJob, 0)
	for agentId, agent := range agentIds {
		//1 job for 1 machine. Each job is named after an agent ID
		target = append(target, TargetJob{
			Labels: Labels{Job: agentId},
			Target: []string{
				//cAdvisor
				agent.InternalIP + ":8080",
				//Node exporter
				agent.InternalIP + ":9100",
			},
		})
	}
	bytes, _ := json.Marshal(target)
	return base64.StdEncoding.EncodeToString(bytes)
}

//Target file for Prometheus service discovery.
//Data structure should be []TargetJob
type TargetJob struct {
	Labels Labels   `json:"labels"`
	Target []string `json:"targets"`
}

type Labels struct {
	Job string `json:"job"`
}
