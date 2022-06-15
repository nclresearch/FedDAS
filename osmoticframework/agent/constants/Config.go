package constants

import (
	"encoding/json"
	"golang.org/x/net/nettest"
	"net"
	"osmoticframework/agent/log"
	"sync"
)

var config configStruct

type configStruct struct {
	RabbitAddress      string   `json:"rabbitAddress"`
	NetworkInterface   string   `json:"networkInterface"`
	DeviceSupport      []string `json:"device_support"`
	SensorSupport      []string `json:"sensor_support"`
	ContainerWhitelist []string `json:"container_whitelist"`
}

func Load(jsonBytes []byte) {
	err := json.Unmarshal(jsonBytes, &config)
	if err != nil {
		log.Fatal.Println("Cannot read properties file")
		log.Fatal.Panicln(err)
	}
}

func GetRabbitAddress() string {
	return config.RabbitAddress
}

var inetMutex sync.Once
var inet string

func GetNetworkInterface() string {
	if config.NetworkInterface != "" {
		return config.NetworkInterface
	}
	inetMutex.Do(func() {
		rif, err := nettest.RoutedInterface("ip", net.FlagUp|net.FlagBroadcast)
		if err != nil {
			log.Fatal.Println("Cannot find default gateway! Is agent connected to a network?")
			log.Fatal.Panicln(err)
		}
		inet = rif.Name
	})
	return inet
}

func GetDeviceSupport() []string {
	return config.DeviceSupport
}

func GetSensorSupport() []string {
	return config.SensorSupport
}

func GetContainerWhitelist() []string {
	return config.ContainerWhitelist
}
