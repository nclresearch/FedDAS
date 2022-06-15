package vars

import (
	"encoding/json"
	"net"
	"osmoticframework/controller/log"
	"regexp"
	"sync"
)

//Stores all of the properties written in the properties file
var config configStruct
var cloudIP string

type configStruct struct {
	RabbitAddress     string   `json:"rabbitAddress"`
	DatabaseAddress   string   `json:"databaseAddress"`
	PrometheusAddress string   `json:"prometheusAddress"`
	KubeConfigPath    string   `json:"kuberConfigPath"`
	EnableProfiler    bool     `json:"enable_profiler,omitempty"`
	ProfilerPort      int      `json:"profiler_port,omitempty" default:"6060"`
	CIRepo            []string `json:"ci_repo,omitempty"`
}

func LoadConfig(jsonBytes []byte) {
	//Unmarshal the json file to a config struct
	err := json.Unmarshal(jsonBytes, &config)
	if err != nil {
		log.Fatal.Println("Cannot read properties file")
		log.Fatal.Fatalln(err)
	}
	log.Info.Println("Registered networks")
	log.Info.Println("Kubernetes config path: " + config.KubeConfigPath)
}

func GetRabbitAddress() string {
	return config.RabbitAddress
}

func GetDatabaseAddress() string {
	return config.DatabaseAddress
}

func GetPrometheusAddress() string {
	if config.PrometheusAddress == "" {
		log.Warn.Println("No Prometheus address specified. Cloud monitoring functions calls will panic!")
	}
	return config.PrometheusAddress
}

func GetKuberConfigPath() string {
	if config.KubeConfigPath == "" {
		log.Warn.Println("No kubernetes config path specified. Cloud monitoring functions calls will panic!")
	}
	return config.KubeConfigPath
}

var cloudIPMutex sync.Once

func GetListeningIP() string {
	cloudIPMutex.Do(func() {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal.Println("Cannot get listening IP")
			log.Fatal.Fatalln(err)
		}
		defer conn.Close()
		cloudIP, _, err = net.SplitHostPort(conn.LocalAddr().String())
		if err != nil {
			log.Fatal.Println("Cannot get listening IP")
			log.Fatal.Fatalln(err)
		}
	})
	return cloudIP
}

func GetRabbitHost() string {
	regex := regexp.MustCompile(`amqps?://\w+:\w+@([a-zA-Z0-9-.]+)(:\d{2,5})?`)
	matches := regex.FindStringSubmatch(GetRabbitAddress())
	return matches[1]
}

func GetRabbitPort() string {
	regex := regexp.MustCompile(`amqps?://\w+:\w+@([a-zA-Z0-9-.]+)(:(\d{2,5}))?`)
	matches := regex.FindStringSubmatch(GetRabbitAddress())
	return matches[3]
}

func GetMysqlHost() string {
	regex := regexp.MustCompile(`\w+:\w+@tcp\(([a-zA-Z0-9.-]+):(\d{2,5})\)/agents`)
	matches := regex.FindStringSubmatch(GetDatabaseAddress())
	return matches[1]
}

func GetMysqlPort() string {
	regex := regexp.MustCompile(`\w+:\w+@tcp\(([a-zA-Z0-9.-]+):(\d{2,5})\)/agents`)
	matches := regex.FindStringSubmatch(GetDatabaseAddress())
	return matches[2]
}

func IsProfilerEnable() bool {
	return config.EnableProfiler
}

func GetProfilerPort() int {
	return config.ProfilerPort
}

func GetCIRepo() []string {
	return config.CIRepo
}
