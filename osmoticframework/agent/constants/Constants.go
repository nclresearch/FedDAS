package constants

import (
	"encoding/base64"
	"fmt"
	"net"
	"osmoticframework/agent/log"
	"osmoticframework/agent/types"
	"runtime"
	"strings"
)

//Constant prometheus configuration
const prometheusConfig = `global:
  scrape_interval: 5s
  evaluation_interval: 5s
scrape_configs:
  - job_name: 'osmotic-prometheus'
    static_configs:
      - targets:
        - 'localhost:9090'
  - job_name: 'osmotic-prometheus-node'
    static_configs:
      - targets:
        - '%s:9100'
  - job_name: 'osmotic-prometheus-cadvisor'
    static_configs:
      - targets:
        - '%s:8080'
`

//All containers that is deployed without controller interaction.
//Only add entries here if the container has to be deployed on all agents.
var DefaultContainers = [...]types.DeployArgs{
	{
		Image: func() string {
			if runtime.GOARCH == "amd64" {
				return "gcr.io/cadvisor/cadvisor:latest"
			} else if runtime.GOARCH == "arm64" || runtime.GOARCH == "arm" || runtime.GOARCH == "386" || runtime.GOARCH == "ppc64le" {
				// cAdvisor is only officially supported on amd64 only
				// This unofficial one provides support for arm64, arm, 386, ppc64le
				return "zcube/cadvisor:latest"
			} else {
				log.Fatal.Panicln("Unsupported architecture: " + runtime.GOARCH)
				return ""
			}
		}(),
		ExposePorts: []types.ExposePort{
			{
				HostPort:      8080,
				ContainerPort: 8080,
			},
		},
		Volumes: []types.Volume{
			{
				HostPath:      "/",
				ContainerPath: "/rootfs",
				ReadOnly:      true,
			},
			{
				HostPath:      "/var/run",
				ContainerPath: "/var/run",
				ReadOnly:      true,
			},
			{
				HostPath:      "/sys",
				ContainerPath: "/sys",
				ReadOnly:      true,
			},
			{
				HostPath:      "/var/lib/docker",
				ContainerPath: "/var/lib/docker",
				ReadOnly:      true,
			},
			{
				HostPath:      "/dev/disk",
				ContainerPath: "/dev/disk",
				ReadOnly:      true,
			},
		},
		RestartPolicy: types.RestartOnFailure,
	},
	{
		Image: "prom/node-exporter",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      9100,
				ContainerPort: 9100,
			},
		},
		Command: []string{
			"--path.procfs=/host/proc",
			"--path.rootfs=/rootfs",
			"--path.sysfs=/host/sys",
			"--collector.filesystem.ignored-mount-points=^/(sys|proc|dev|host|etc)($$|/)",
		},
		Volumes: []types.Volume{
			{
				HostPath:      "/proc",
				ContainerPath: "/host/proc",
				ReadOnly:      true,
			},
			{
				HostPath:      "/sys",
				ContainerPath: "/host/sys",
				ReadOnly:      true,
			},
			{
				HostPath:      "/",
				ContainerPath: "/rootfs",
				ReadOnly:      true,
			},
		},
		RestartPolicy: types.RestartOnFailure,
	},
}

func PrometheusContainer() types.DeployArgs {
	container := types.DeployArgs{
		Image: "prom/prometheus",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      9090,
				ContainerPort: 9090,
			},
		},
		//Prometheus requires a configuration FILE.
		//To get around this, we can first put the contents of the config file inside of an environment variable
		//However, environment variables cannot contain backslash escapes. Corrupting whitespace aware format files like YAML
		//We can base64 encode the file instead. Unfortunately, not all containers comes with busybox/coreutils.
		//Containers with long arguments may also fail due to an argument size limit imposed by the shell.
		//This is the best method for now.
		Entrypoint: []string{"/bin/sh", "-c", "echo $PROM_CONFIG | base64 -d > /etc/prometheus/prometheus.yml && /bin/prometheus " +
			//Default parameters of Prometheus
			"--config.file=/etc/prometheus/prometheus.yml " +
			"--storage.tsdb.path=/prometheus " +
			"--web.console.libraries=/usr/share/prometheus/console_libraries " +
			"--web.console.templates=/usr/share/prometheus/consoles",
		},
		Environment: []types.Environment{
			{
				Name:  "PROM_CONFIG",
				Value: PrometheusConfig(),
			},
		},
		GPU:           nil,
		RestartPolicy: types.RestartOnFailure,
	}
	return container
}

func PrometheusConfig() string {
	// Attempt to get the local IP of the host
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Fatal.Fatalln("Could not get local IP:", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()
	localIP := strings.Split(conn.LocalAddr().String(), ":")[0]
	yml := fmt.Sprintf(prometheusConfig, localIP, localIP)
	encodedStr := base64.StdEncoding.EncodeToString([]byte(yml))
	return encodedStr
}
