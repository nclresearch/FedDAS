package types

import (
	"time"
)

type AlertType string

const (
	AlertContainerCrash AlertType = "containerCrash"
)

type OfflineAgent struct {
	ID    string
	Agent Agent
}

type FailedRequest struct {
	ID      string
	ToAgent string
	API     string
	Command string
	Ack     bool
	Time    time.Time
	Content interface{}
	Timeout float64
	Error   string
}

type DatabaseErrorReport struct {
	Query     string
	QueryArgs []string
	Error     error
}

type ContainerCrashReport struct {
	AgentId  string
	ID       string
	Name     string
	Image    string
	Status   string
	ExitCode int
}

type PerformanceAlertType string

const (
	PerformanceAlertTypeCPUUsageHigh   PerformanceAlertType = "cpu-usage-high"
	PerformanceAlertTypeMemoryPressure PerformanceAlertType = "memory-pressure"
	PerformanceAlertTypeNetworkLatency PerformanceAlertType = "network-latency"
	PerformanceAlertTypeDiskPressure   PerformanceAlertType = "disk-pressure"
)

type PerformanceReport struct {
	AgentId   string
	CloudSide bool
	Container string
	AlertType PerformanceAlertType
}
