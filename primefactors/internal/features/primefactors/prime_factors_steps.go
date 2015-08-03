package primefactors

import (
	"log"
	. "github.com/lsegal/gucumber"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"os"
	"fmt"
	"net/http"
	"strings"
)

func readDockerEnv() (string,string) {

	dockerHome := os.Getenv("DOCKER_HOST")
	if dockerHome == "" {
		log.Fatal("DOCKER_HOST environment variable not set.")
	}

	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	if dockerCertPath == "" {
		log.Fatal("DOCKER_CERT_PATH environment variable not set.")
	}

	return dockerHome, dockerCertPath
}

func buildDockerTLSConfig(dockerCertPath string) *tls.Config {


	caFile := fmt.Sprintf("%s/ca.pem", dockerCertPath)
	certFile := fmt.Sprintf("%s/cert.pem", dockerCertPath)
	keyFile := fmt.Sprintf("%s/key.pem", dockerCertPath)

	tlsConfig := &tls.Config{}

	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	pemCerts, _ := ioutil.ReadFile(caFile)

	tlsConfig.RootCAs       = x509.NewCertPool()
	tlsConfig.ClientAuth    = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates  = []tls.Certificate{cert}

	tlsConfig.RootCAs.AppendCertsFromPEM(pemCerts)

	return tlsConfig
}

func getAcceptanceTestContainerInfo(docker *dockerclient.DockerClient, containerType string) *dockerclient.ContainerInfo {

	// Get only running containers
	containers, err := docker.ListContainers(false, false, "")
	if err != nil {
		log.Fatal(err)
	}

	//Loop through them until we find a match
	for _, c := range containers {
		xtContainerType, ok := c.Labels["xt-container-type"]
		if ok && xtContainerType == containerType {
			//Grab the information for the container
			info, err := docker.InspectContainer(c.Id)
			if err != nil {
				log.Fatal(err)
			}

			return info
		}
	}

	return nil
}

type TestContext struct {
	input int
	outputStatus int
	outputData []byte
}

type ContainerContext struct {
	ImageName string
	Labels map[string]string
}

func createAndStartContainer(docker *dockerclient.DockerClient, ctx *ContainerContext) string {

	containerConfig := dockerclient.ContainerConfig{
		Image: ctx.ImageName,
		Labels:ctx.Labels,
		ExposedPorts:map[string]struct{}{
			"3000/tcp":{},
		},
	}

	var err error
	containerId, err := docker.CreateContainer(&containerConfig, "")
	if err != nil {
		log.Fatal(err)
	}

	pb := dockerclient.PortBinding{HostPort:"8080"}
	portBindings := []dockerclient.PortBinding{pb}
	hostConfig := &dockerclient.HostConfig{
		PortBindings:map[string][]dockerclient.PortBinding {
			"3000/tcp":portBindings,
		},
	}
	err = docker.StartContainer(containerId, hostConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("...we just started a docker container...")

	return containerId
}

func requiredImageAvailable(docker *dockerclient.DockerClient, name string) (bool,error) {
	images, err := docker.ListImages(true)
	if err != nil {
		return false, err
	}

	for _, i := range images {
		for _,t := range i.RepoTags {
			if strings.Index(t, name) == 0 {
				return true, nil
			}
		}
	}

	return false, nil
}

func init() {

	var containerId string
	var docker *dockerclient.DockerClient
	var testContext *TestContext

	Before("@primefactors", func() {
		//Grab the environment
		dockerHost, dockerCertPath := readDockerEnv()

		// Init the client
		log.Println("Create docker client")
		docker, _ = dockerclient.NewDockerClient(dockerHost, buildDockerTLSConfig(dockerCertPath))

		// Is the container already running?
		log.Println("Check to see if container is already started")
		info := getAcceptanceTestContainerInfo(docker, "atest")
		if info != nil {
			log.Println("Container found - state is: ", info.State.StateString())
			log.Println("Test is ready to run")
			return
		}

		// If the container is not running, is the image required for the test present such that we
		// can create a container based on the required image?
		log.Println("Container not running - create container context")
		containerCtx := ContainerContext{
			ImageName:"pfservice",
		}

		containerCtx.Labels = make(map[string]string)
		containerCtx.Labels["xt-container-type"] = "atest"

		log.Println("Verify required image is present")
		imagePresent, err := requiredImageAvailable(docker, containerCtx.ImageName)
		if err != nil {
			log.Fatal(err)
		}

		if !imagePresent {
			log.Fatal("Cannot run test as required image (", containerCtx.ImageName, ") is not present.")
		}

		//Create and start the container.
		log.Println("Create and start the container")
		containerId = createAndStartContainer(docker, &containerCtx)

	})

	Given(`^A prime factor resource value of (\d+)$`, func(n int) {
		testContext = new(TestContext)
		testContext.input = n
	})

	When(`^I call the prime factors service$`, func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:8080/pf/%d", testContext.input))
		if err != nil {
			T.Errorf("Call to endpoint has failed: ", err.Error())
		}

		testContext.outputData,_ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		testContext.outputStatus = resp.StatusCode
	})

	Then(`^The prime factors for the resouce value are returned$`, func() {
		if testContext.outputStatus != http.StatusOK {
			T.Errorf("Invalid status returned, expecting 200 OK: ", testContext.outputStatus)
		}

		outputString := string(testContext.outputData)
		if string(testContext.outputData) != "[5,5,5,5,13]" {
			T.Errorf("Unexpected output - expected [5,5,5,5,13], got ", outputString)
		}
	})

	After("@primefactors", func() {
		if containerId != "" {
			log.Println("stop container")
			docker.StopContainer(containerId, 5)
			log.Println("remove container")
			docker.RemoveContainer(containerId, false, false)
		}
	})

}