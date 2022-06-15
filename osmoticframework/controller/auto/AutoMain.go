package auto

import (
	"github.com/docker/docker/pkg/fileutils"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"os"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
	"osmoticframework/controller/util"
	"osmoticframework/controller/vars"
	"path"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var Events = make(chan string)
var AgentJoin = make(chan string)
var startLoop = make(chan bool)

var repoDirectories sync.Map

var executorCount int64
var executorCountMax int64
var tmpDir string

// AutoMain Auto deploy logic main
func AutoMain() {
	//Wait for controller API to get ready
	for vars.GetApiReady() {
	}
	once()
	//Start loop
	go func() {
		<-startLoop
		log.Info.Println("Starting CI loop")
		for {
			loop()
		}
	}()
	//Listening to events
	go func() {
		for e := range Events {
			customIn(e)
		}
	}()
	go func() {
		for e := range AgentJoin {
			agentJoin(e)
		}
	}()
	go func() {
		//Crash handler
		for crash := range alert.ContainerCrash {
			crashHandle(crash)
		}
	}()
	//Disconnected agents
	for agent := range alert.AgentDisconnect {
		disconnectedAgent(agent)
	}
	//Performance issues
	for agent := range alert.PerformanceIssues {
		performanceAlert(agent)
	}
}

//Deploy logic that only run once
func once() {
	err := enableCrossBuild()
	if err != nil {
		log.Fatal.Println("Error on enable cross build: " + err.Error())
	}
	tmpDir, err = ioutil.TempDir("", "osmotic")
	if err != nil {
		log.Fatal.Println("Error on create tmp dir: " + err.Error())
	}
	err = request.KCreateConfigMap(influxdbConfig)
	if err != nil {
		log.Error.Println("Error creating config map: " + err.Error())
	}
	err = request.KRunDeployment(influxdbDeployment, nil)
	if err != nil {
		log.Error.Println("Error on deploy influxdb pod" + err.Error())
	}
	_, err = request.KCreateService(influxdbService)
	if err != nil {
		log.Error.Println("Error on deploy influxdb service" + err.Error())
	}
	err = request.KCreateConfigMap(mqttConfig)
	if err != nil {
		log.Error.Println("Error creating config map: " + err.Error())
	}
	err = request.KRunDeployment(mqttDeployment, nil)
	if err != nil {
		log.Error.Println("Error on deploy mqtt pod" + err.Error())
	}
	_, err = request.KCreateService(mqttService)
	if err != nil {
		log.Error.Println("Error on deploy mqtt service" + err.Error())
	}
	err = request.KRunDeployment(aggregatorDeployment, nil)
	if err != nil {
		log.Error.Println("Error on deploy aggregator pod" + err.Error())
	}
	_, err = request.KCreateService(aggregatorService)
	if err != nil {
		log.Error.Println("Error on deploy aggregator service" + err.Error())
	}
	close(startLoop)
}

//Deploy logic that runs in a loop
func loop() {
	//Atom feeds in Github can only be read via basic auth. So for authentication you need a Github token.
	//Read github credentials - You can change the credentials in the file and it will be used in the next run
	//Note: GitHub now requires a token to for the command line interface. The token is treated as the password.
	//githubCredential, err := ioutil.ReadFile(vars.GetCredDirectory() + "/github.txt")
	//if err != nil {
	//	log.Fatal.Fatalln("Error on read github credential file: " + err.Error())
	//}
	//Base64 encode credentials
	for _, repo := range vars.GetCIRepo() {
		//Check if repo is already in filesystem

		log.Info.Println("Checking for updates on " + repo)
		repoName := strings.Split(repo, "/")[len(strings.Split(repo, "/"))-1]
		_, ok := repoDirectories.Load(repo)
		if ok {
			//Open repo directory and pull
			log.Info.Println("Pulling updates on " + repo)
			r, err := git.PlainOpen(path.Join(tmpDir, repoName))
			if err != nil {
				log.Error.Println("Error on open repo: " + err.Error())
				continue
			}
			w, err := r.Worktree()
			if err != nil {
				log.Error.Println("Error on open worktree: " + err.Error())
				continue
			}
			err = w.Pull(&git.PullOptions{
				RemoteName: "origin",
				//Auth: &gitHttp.BasicAuth{
				//	Username: strings.Split(string(githubCredential), ":")[0],
				//	Password: strings.Split(string(githubCredential), ":")[1],
				//},
			})
			if err != nil {
				if err == git.NoErrAlreadyUpToDate {
					log.Info.Println("No updates on " + repo)
				} else {
					log.Error.Println("Error on pull: " + err.Error())
				}
				continue
			}
		} else {
			//Clone repo
			_, err := git.PlainClone(tmpDir+"/"+repoName, false, &git.CloneOptions{
				URL: repo,
				//Auth: &gitHttp.BasicAuth{
				//	Username: strings.Split(string(githubCredential), ":")[0],
				//	Password: strings.Split(string(githubCredential), ":")[1],
				//},
			})
			if err != nil {
				log.Error.Println("Error on clone: " + err.Error())
				continue
			}
			repoDirectories.Store(repo, tmpDir+"/"+repoName)
		}
		// Perform tasks on repo
		switch repoName {
		case "CIKM22_DEMO":
			//Check how many executors to deploy from config file
			flConfig, err := os.Open(path.Join(path.Join(tmpDir, repoName), "ubifl_server/experiments/configs/cifar10/conf.yml"))
			if err != nil {
				log.Error.Println("Error on open config file: " + err.Error())
				break
			}
			yamlConfig, err := ioutil.ReadAll(flConfig)
			if err != nil {
				log.Error.Println("Error on read config file: " + err.Error())
				flConfig.Close()
				break
			}
			var config map[string]interface{}
			err = yaml.Unmarshal(yamlConfig, &config)
			if err != nil {
				log.Error.Println("Error on unmarshal config file: " + err.Error())
				flConfig.Close()
				break
			}
			//Run this on bare metal. Not in a container!
			// Copy pth file if not exist
			if _, err := os.Stat(path.Join(tmpDir, repoName, "util_components/NAS/NAS_Bench_201/nas_201_pth/NAS-Beach-201-v1_1-096897.pth")); os.IsNotExist(err) {
				err := os.MkdirAll(path.Join(tmpDir, "CIKM22_DEMO/util_components/NAS/NAS_Bench_201/nas_201_pth"), 0755)
				if err != nil {
					log.Error.Println("Error on create directory: " + err.Error())
				}
				_, err = fileutils.CopyFile("/home/campus.ncl.ac.uk/nwhs3/CIKM22_FL/util_components/NAS/NAS_Bench_201/nas_201_pth/NAS-Bench-201-v1_1-096897.pth", path.Join(tmpDir, "CIKM22_DEMO/util_components/NAS/NAS_Bench_201/nas_201_pth/NAS-Bench-201-v1_1-096897.pth"))
				if err != nil {
					log.Error.Println("Error on copy nas_201_pth: " + err.Error())
				}
			}
			atomic.StoreInt64(&executorCountMax, int64(config["device"].(map[string]interface{})["num_of_clients"].(float64)))
			log.Info.Println("Number of required executors is now " + strconv.FormatInt(atomic.LoadInt64(&executorCountMax), 10))
			flConfig.Close()
			// Begin build image
			log.Info.Println("Building image for " + repoName)
			client, err := docker.NewClientFromEnv()
			if err != nil {
				log.Error.Println("Error on getting docker client context: " + err.Error())
				continue
			}
			log.Info.Println("Building aggregator image")
			// Build aggregator
			err = client.BuildImage(docker.BuildImageOptions{
				Name:         "localhost:32000/iot_aggregator",
				Dockerfile:   "ubifl_server/CI/cpu_server/Dockerfile_aggregator",
				ContextDir:   path.Join(tmpDir, repoName),
				OutputStream: os.Stdout,
			})
			if err != nil {
				log.Error.Println("Error on building image iot_aggregator: " + err.Error())
				continue
			}
			log.Info.Println("Building executor image")
			// Build executor (arm64)
			// This forces the build to use the arm64 architecture, which runs inside a static qemu emulator
			err = client.BuildImage(docker.BuildImageOptions{
				Name:         "localhost:32000/iot_executor",
				Dockerfile:   "ubifl_edge/CI/jetson_nano/Dockerfile_executor",
				ContextDir:   path.Join(tmpDir, repoName),
				Platform:     "arm64",
				OutputStream: os.Stdout,
			})
			if err != nil {
				log.Error.Println("Error on building image iot_executor: " + err.Error())
				continue
			}
			// Push images to registry
			// Authorization is still required even for unauthenticated registries, use "docker" as username and empty password1k
			log.Info.Println("Pushing aggregator image to registry")
			err = client.PushImage(docker.PushImageOptions{
				Name: "localhost:32000/iot_aggregator",
				Tag:  "latest",
			}, docker.AuthConfiguration{
				Username: "docker",
				Password: "",
			})
			if err != nil {
				log.Error.Println("Error on pushing image iot_aggregator: " + err.Error())
				continue
			}
			log.Info.Println("Pushing executor image to registry")
			err = client.PushImage(docker.PushImageOptions{
				Name: "localhost:32000/iot_executor",
				Tag:  "latest",
			}, docker.AuthConfiguration{
				Username: "docker",
				Password: "",
			})
			if err != nil {
				log.Error.Println("Error on pushing image iot_executor: " + err.Error())
				continue
			}
			//Refresh all current containers
			log.Info.Println("Refreshing all containers")
			err = request.KUpdateDeployment(aggregatorDeployment, nil)
			if err != nil {
				log.Error.Println("Failed updating aggregator: " + err.Error())
			}
			vars.Agents.Range(func(_agentId, _agent interface{}) bool {
				agentId := _agentId.(string)
				req := request.ListRequest(agentId, 10)
				resp := <-req.Result
				containers := resp.Content.([]types.Container)
				for _, container := range containers {
					if container.Image == "19scomps001.ncl.ac.uk/iot_executor" {
						updateReq := request.UpdateRequest(agentId, container.ID, executorContainer, types.AuthInfo{}, 120)
						updateResp := <-updateReq.Result
						if updateResp.ResultType == request.Error {
							log.Error.Println("Failed updating container on agent " + agentId + ": " + err.Error())
						}
						break
					}
				}
				return true
			})
		}
	}
	time.Sleep(time.Second * 10)
}

//Custom event fired from outside containers
//To do this, containers should connect to the queue "event" and send messages here
func customIn(eventMessage string) {
	//Insert auto deploy logic here
	log.Info.Println(eventMessage)
}

//Agent join event
func agentJoin(agentId string) {
	//Insert logic when new agent joined
	if executorCount == atomic.LoadInt64(&executorCountMax) {
		return
	}
	log.Info.Println("Deploying executor to agent: " + agentId)
	response := request.RunRequest(agentId, executorContainer, types.AuthInfo{}, 500)
	result := <-response.Result
	if result.ResultType == request.Error {
		log.Error.Println("Error on deploy executor to agent: " + agentId + " - " + result.Content.(error).Error())
	}
	log.Info.Println("Deploying influxdb to agent: " + agentId)
	response = request.RunRequest(agentId, influxdbContainer, types.AuthInfo{}, 180)
	result = <-response.Result
	if result.ResultType == request.Error {
		log.Error.Println("Error on deploy influxdb to agent: " + agentId + " - " + result.Content.(error).Error())
	}
	executorCount++
}

//Container crash handler. This include s succeeded containers
//Crashes are automatically logged. You do not need to log again
func crashHandle(crash types.ContainerCrashReport) {
	//Ignore any crashes that occurred during the first few seconds of controller startup
	//As containers might require other dependencies containers to start, if the dependency hasn't started yet. They will just crash
	//This gives a grace period before the controller really treat container crashes as real issues
	if util.Uptime() < time.Second*30 {
		return
	}
}

func disconnectedAgent(agent types.OfflineAgent) {
	//Handle agents that have disconnected.
	//They usually have containers running while it disconnects, we must handle any workflow that might have been interrupted by this event
	executorCount--
	log.Error.Printf("Affected container IDs: %#v", agent.Agent.Containers)
}

func performanceAlert(alert types.PerformanceReport) {
	//Handle performance alerts
	//Alerts are automatically logged
}

func enableCrossBuild() error {
	//Enable cross build in docker
	//This forces image builds in different architectures to start inside a qemu container
	log.Info.Println("Enabling multiarch build capabilities")
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	qemuContainer, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: "multiarch/qemu-user-static",
			Cmd:   []string{"--reset", "-p", "yes"},
		},
		HostConfig: &docker.HostConfig{
			Privileged: true,
			AutoRemove: true,
		},
	})
	if err != nil {
		return err
	}
	err = client.StartContainer(qemuContainer.ID, nil)
	if err != nil {
		return err
	}
	return nil
}
