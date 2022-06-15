package auto

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"osmoticframework/controller/types"
	"osmoticframework/controller/util"
)

//Insert container templates here
var influxdbDeployment = types.KDeployArgs{
	DeploymentName: "influxdb",
	Replicas:       1,
	Strategy:       v1.RecreateDeploymentStrategyType,
	PodArgs: types.KPodArgs{
		Image: "influxdb",
		Label: map[string]string{
			"app": "influxdb",
		},
		Name: "influxdb",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      8086,
				ContainerPort: 8086,
				Protocol:      "TCP",
			},
		},
		Environment: []types.Environment{
			{
				Name:  "DOCKER_INFLUXDB_INIT_MODE",
				Value: "setup",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_USERNAME",
				Value: "admin",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_PASSWORD",
				Value: "admin123",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_ORG",
				Value: "NCL",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_BUCKET",
				Value: "pre_setting",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_RETENTION",
				Value: "30d",
			},
			{
				Name:  "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN",
				Value: "TokenForNCLProject",
			},
		},
		Volumes: []types.KVolume{
			{
				MountPoint: "/docker-entrypoint-initdb.d",
				Type:       types.TypeConfigMap,
				ConfigMap: &types.ConfigMapVolume{
					Name: "influxdb-config",
					Mode: util.Int32Ptr(0777),
				},
			},
		},
		PullPolicy:    corev1.PullAlways,
		RestartPolicy: corev1.RestartPolicyAlways,
	},
}

var influxdbConfig = types.KConfigMap{
	Name: "influxdb-config",
	Data: map[string]string{
		"init.sh": `#!/bin/sh
set -x
influx bucket create -n running_monitoring -r 30d
`,
	},
}

var influxdbService = types.KServiceArgs{
	Name: "influxdb-service",
	Selector: map[string]string{
		"app": "influxdb",
	},
	Ports: []types.KServicePort{
		{
			Name:       "influxdb-port",
			Host:       8086,
			TargetPort: 8086,
			NodePort:   30004,
			Protocol:   "TCP",
		},
	},
	Type: corev1.ServiceTypeNodePort,
}

var mqttDeployment = types.KDeployArgs{
	PodArgs: types.KPodArgs{
		Image: "eclipse-mosquitto",
		Label: map[string]string{
			"app": "mqtt",
		},
		Name: "mqtt",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      1883,
				ContainerPort: 1883,
				Protocol:      "TCP",
			},
			{
				HostPort:      9001,
				ContainerPort: 9001,
				Protocol:      "TCP",
			},
		},
		Volumes: []types.KVolume{
			{
				MountPoint: "/mosquitto/config",
				Type:       types.TypeConfigMap,
				ConfigMap: &types.ConfigMapVolume{
					Name: "mosquitto-config",
					Mode: util.Int32Ptr(0644),
				},
			},
		},
		PullPolicy:    corev1.PullAlways,
		RestartPolicy: corev1.RestartPolicyAlways,
	},
	DeploymentName: "mqtt",
	Replicas:       1,
	Strategy:       v1.RecreateDeploymentStrategyType,
}

var mqttConfig = types.KConfigMap{
	Name: "mosquitto-config",
	Data: map[string]string{
		"mosquitto.conf": `
port 1883
allow_anonymous true
`,
	},
}

var mqttService = types.KServiceArgs{
	Name: "mqtt-service",
	Ports: []types.KServicePort{
		{
			Name:       "mqtt-port",
			Host:       1883,
			TargetPort: 1883,
			NodePort:   30001,
			Protocol:   "TCP",
		},
		{
			Name:       "mqtt-ws-port",
			Host:       9001,
			TargetPort: 9001,
			NodePort:   30002,
			Protocol:   "TCP",
		},
	},
	Selector: map[string]string{
		"app": "mqtt",
	},
	Type: corev1.ServiceTypeNodePort,
}

var aggregatorService = types.KServiceArgs{
	Name: "aggregator-service",
	Ports: []types.KServicePort{
		{
			Name:       "aggregator-port",
			Host:       8088,
			TargetPort: 8088,
			NodePort:   30003,
			Protocol:   "TCP",
		},
	},
	Selector: map[string]string{
		"iot": "aggregator",
	},
	Type: corev1.ServiceTypeNodePort,
}

var aggregatorDeployment = types.KDeployArgs{
	PodArgs: types.KPodArgs{
		Image: "localhost:32000/iot_aggregator",
		Label: map[string]string{
			"iot": "aggregator",
		},
		Name: "aggregator",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      8088,
				ContainerPort: 8088,
				Protocol:      "TCP",
			},
		},
		PullPolicy:    corev1.PullAlways,
		RestartPolicy: corev1.RestartPolicyAlways,
	},
	DeploymentName: "aggregator",
	Replicas:       1,
	Strategy:       v1.RecreateDeploymentStrategyType,
}

var executorContainer = types.DeployArgs{
	Image: "19scomps001.ncl.ac.uk:32000/iot_executor",
	GPU: &types.GPU{
		Count: util.Int64Ptr(-1),
	},
	RestartPolicy: types.RestartOnFailure,
	PullOptions:   types.PullAlways,
}

var influxdbContainer = types.DeployArgs{
	Image: "influxdb",
	ExposePorts: []types.ExposePort{
		{
			HostPort:      8086,
			ContainerPort: 8086,
			Protocol:      "TCP",
		},
	},
	Entrypoint: []string{"/bin/bash", "-c", "$INIT_SCRIPT | base64 -d > /docker-entrypoint-initdb.d/init.sh && /entrypoint.sh"},
	Environment: []types.Environment{
		{
			Name:  "DOCKER_INFLUXDB_INIT_MODE",
			Value: "setup",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_USERNAME",
			Value: "admin",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_PASSWORD",
			Value: "admin123",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_ORG",
			Value: "NCL",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_BUCKET",
			Value: "pre_setting",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_RETENTION",
			Value: "30d",
		},
		{
			Name:  "DOCKER_INFLUXDB_INIT_ADMIN_TOKEN",
			Value: "TokenForNCLProject",
		},
		{
			Name: "INIT_SCRIPT",
			/*
				#!/bin/sh
				set -x
				influx bucket create -n running_monitoring -r 30d
			*/
			Value: "IyEvYmluL3NoCnNldCAteAppbmZsdXggYnVja2V0IGNyZWF0ZSAtbiBydW5uaW5nX21vbml0b3JpbmcgLXIgMzBkCg==",
		},
	},
	RestartPolicy: types.RestartOnFailure,
	PullOptions:   types.PullAlways,
}
