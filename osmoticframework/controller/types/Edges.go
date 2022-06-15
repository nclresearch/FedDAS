package types

//Edge object to store all connected edge devices

type Agent struct {
	InternalIP    string
	DeviceSupport []string
	SensorSupport []string
	Containers    []string //Array of container IDs. The list of containers hosted in the agent device
	LastAlive     int64    //Last time the agent sent a heartbeat. In UNIX nanosecond timestamp.
	PingSeq       int64    //Last ping sequence number
}
