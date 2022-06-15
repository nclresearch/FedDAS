package alert

import "osmoticframework/controller/types"

//The alert system
//The controller will fire a notification (A message to a channel) in the event any critical events occurred

// AgentDisconnect Returns the agent ID that has disconnected.
var AgentDisconnect = make(chan types.OfflineAgent)

// DatabaseErrors Database errors
var DatabaseErrors = make(chan types.DatabaseErrorReport)

// ContainerCrash Container crash
var ContainerCrash = make(chan types.ContainerCrashReport)

// PerformanceIssues Performace issues
var PerformanceIssues = make(chan types.PerformanceReport)
