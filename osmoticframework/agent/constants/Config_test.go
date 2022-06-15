package constants

import (
	_ "osmoticframework/agent/log"
	"testing"
)

func TestConfig(t *testing.T) {
	const config = `
{
  "rabbitAddress": "amqp://guest:guest@localhost:5672",
  "networkInterface": "wlan0",
  "device_support": [
	"gpu"
  ],
  "sensor_support": [
	"camera"
  ],
  "container_whitelist": [
	"mysql",
	"rabbitmq:management"
  ]
}`
	Load([]byte(config))
	const expectRabbitAddr = "amqp://guest:guest@localhost:5672"
	rabbitAddr := GetRabbitAddress()
	if rabbitAddr != expectRabbitAddr {
		t.Errorf("Rabbit address incorrect. Got %s, Want %s", rabbitAddr, expectRabbitAddr)
	}
	const expectNetworkInterface = "wlan0"
	netInterface := GetNetworkInterface()
	if netInterface != expectNetworkInterface {
		t.Errorf("Network interface incorrect. Got %s, Want %s", netInterface, expectNetworkInterface)
	}
	const expectDev = "gpu"
	device := GetDeviceSupport()
	length := len(device)
	if length != 1 {
		t.Errorf("Device support length incorrect. Got %d, Want %d", length, 1)
	}
	if device[0] != expectDev {
		t.Errorf("Device support incorrect. Got %s, Want %s", device[0], expectDev)
	}
	const expectSensor = "camera"
	sensor := GetSensorSupport()
	length = len(sensor)
	if length != 1 {
		t.Errorf("Device support length incorrect. Got %d, Want %d", length, 1)
	}
	if sensor[0] != expectSensor {
		t.Errorf("Device support incorrect. Got %s, Want %s", sensor[0], expectSensor)
	}
	var expectWhitelist = [...]string{"mysql", "rabbitmq:management"}
	whitelist := GetContainerWhitelist()
	length = len(whitelist)
	if length != 2 {
		t.Errorf("Container whitelist length incorrect. Got %d, Want %d", length, 2)
	}
	if whitelist[0] != expectWhitelist[0] {
		t.Errorf("Container whitelist incorrect. Got %s, Want %s", whitelist[0], expectWhitelist[0])
	}
	if whitelist[1] != expectWhitelist[1] {
		t.Errorf("Container whitelist incorrect. Got %s, Want %s", whitelist[1], expectWhitelist[1])
	}

}
