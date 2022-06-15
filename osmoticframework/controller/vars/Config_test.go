package vars

import (
	_ "osmoticframework/controller/log"
	"regexp"
	"testing"
)

func TestConfig(t *testing.T) {
	const config = `
{
  "rabbitAddress": "amqp://guest:guest@localhost:5672",
  "databaseAddress": "root:root@tcp(localhost:3306)/agents",
  "prometheusAddress": "https://192.168.1.1:9090/api/v1",
  "kuberConfigPath": "/etc/kubernetes/admin.conf",
  "listeningInterface": "eth0"
}
`
	LoadConfig([]byte(config))
	const expectRabbitAddr = "amqp://guest:guest@localhost:5672"
	rabbitAddr := GetRabbitAddress()
	if rabbitAddr != expectRabbitAddr {
		t.Errorf("Rabbit address incorrect. Got %s, Want %s", rabbitAddr, expectRabbitAddr)
	}
	const expectDatabase = "root:root@tcp(localhost:3306)/agents"
	databaseAddress := GetDatabaseAddress()
	if databaseAddress != expectDatabase {
		t.Errorf("Network interface incorrect. Got %s, Want %s", databaseAddress, expectDatabase)
	}
	const expectProm = "https://192.168.1.1:9090/api/v1"
	prom := GetPrometheusAddress()
	if prom != expectProm {
		t.Errorf("Device support incorrect. Got %s, Want %s", prom, expectProm)
	}
	const expectKube = "/etc/kubernetes/admin.conf"
	kube := GetKuberConfigPath()
	if kube != expectKube {
		t.Errorf("Kubernetes config path incorrect. Got %s, Want %s", kube, expectKube)
	}
	const ipRegex = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))$`
	ip := GetListeningIP()
	regex := regexp.MustCompile(ipRegex)
	if !regex.MatchString(ip) {
		t.Errorf("Listening IP incorrect. Got %s, Expect format %s", ip, ipRegex)
	}
}
