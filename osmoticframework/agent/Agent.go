package main

import (
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"osmoticframework/agent/api"
	config "osmoticframework/agent/constants"
	deployer "osmoticframework/agent/docker"
	"osmoticframework/agent/log"
	"strings"
)

func main() {
	jsonFile, err := os.Open("properties.json")
	if err != nil {
		log.Fatal.Println("Properties file not found. Please refer to documentation", nil)
		log.Fatal.Fatalln("https://github.com/lukewen427/osmoticframework/wiki/Agent-and-controller")
	}
	log.Info.Println("Loading config file")
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	config.Load(bytes)
	//Prepare prometheus config filepath
	prepareProm()
	//Initialize docker
	err = deployer.Init()
	if err != nil {
		log.Fatal.Println("Failed to initialize docker")
		log.Fatal.Panicln(err)
	}
	//Remove all containers before starting the agent
	//This is to ensure there are no dangling containers running if the agent has crashed and restarted
	stopContainers()
	api.Init()
}

//Prepare prometheus config filepath
//target.json will be used for file based Prometheus service discovery
func prepareProm() {
	tmp := "/tmp"
	log.Info.Println("Preparing files")
	if unix.Access(tmp, unix.W_OK) != nil {
		log.Fatal.Panicln("no write access to /tmp")
	}
	tmp = tmp + "/osmotic-prometheus"
	var err error
	if _, e := os.Stat(tmp); os.IsNotExist(e) {
		err = os.Mkdir(tmp, 0777)
	}
	if err != nil {
		log.Fatal.Println("failed writing to /tmp")
		log.Fatal.Panicln(err)
	}
	if err != nil {
		log.Fatal.Println("failed writing to /tmp")
		log.Fatal.Panicln(err)
	}
}

//Deletes all containers on agent startup, except any whitelisted
func stopContainers() {
	containers, err := deployer.DockerList()
	if err != nil {
		log.Warn.Println("Cannot check Docker status. This may affect performance on the edge device")
		log.Warn.Println(err)
		return
	}
	if len(containers) > 0 {
		log.Info.Println("Cleaning up containers")
	}
	for _, container := range containers {
		//Skip container removal if it's in whitelist.
		skip := false
		for _, whiteList := range config.GetContainerWhitelist() {
			if container.Image == whiteList {
				skip = true
				break
			}
		}
		//Prevent removing itself
		if skip || strings.HasPrefix(container.Image, "osmotic_agent") {
			continue
		}
		log.Info.Printf("Cleaning %s with image %s\n", container.ID, container.Image)
		err = deployer.Stop(container.ID)
		if err != nil {
			if !strings.HasPrefix(err.Error(), "Container not running") {
				log.Error.Printf("Failed stopping container %s. Either the container is not running, or this may affect performance on the edge device\n", container.ID)
				log.Error.Println(err)
			}
		}
		err = deployer.Delete(container.ID, false)
		if err != nil {
			log.Error.Printf("Failed removing container %s. This may affect performance on the edge device\n", container.ID)
			log.Error.Println(err)
		}
	}
}
