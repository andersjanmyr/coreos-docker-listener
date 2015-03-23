package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/andersjanmyr/awsinfo"
	"github.com/coreos/go-etcd/etcd"
)

type Event struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type Config struct {
	Hostname string
}

type NetworkSettings struct {
	IpAddress   string
	PortMapping map[string]map[string]string
}

type Container struct {
	Id              string
	Image           string
	Config          *Config
	NetworkSettings *NetworkSettings
}

type ContainerId struct {
	Id string
}

var client http.Client
var etcdClient *etcd.Client
var hostname string

func init() {
	tr := &http.Transport{
		Dial: fakeDial,
	}
	client = http.Client{Transport: tr}
	etcdClient = etcd.NewClient([]string{
		"http://172.17.42.1:4001", "http://localhost:4001"})
}

func getContainer(id string) ([]byte, error) {
	res, err := client.Get("http://ignor.ed/containers/" + id + "/json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, err
}

func fakeDial(proto, addr string) (conn net.Conn, err error) {
	return net.Dial("unix", "/var/run/docker.sock")
}

func getContainerIds() ([]ContainerId, error) {
	res, err := client.Get("http://ignor.ed/containers/json")
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		d := json.NewDecoder(res.Body)
		var containerIds []ContainerId

		if err = d.Decode(&containerIds); err != nil {
			return nil, err
		}

		return containerIds, nil
	}
	return nil, err
}

func registerContainer(id string, event string) error {
	data, err := getContainer(id)

	var container Container
	err = json.Unmarshal(data, &container)
	if err != nil {
		return err
	}
	if event == "start" {
		if err := registerInEtcd(id, string(data)); err != nil {
			return err
		}
	} else {
		if err := deregisterFromEtcd(id); err != nil {
			return err
		}
	}
	return nil
}

func registerInEtcd(id string, data string) error {
	log.Printf("registerInEtcd %v", id)
	key := "/our/machines/" + hostname + "/containers/" + id
	_, err := etcdClient.Set(key, data, 60)
	return err
}

func deregisterFromEtcd(id string) error {
	log.Printf("deregisterFromEtcd %v", id)
	key := "/our/machines/" + hostname + "/containers/" + id
	_, err := etcdClient.Delete(key, true)
	return err
}

func registerMachine() error {
	var hostname string
	info, err := awsinfo.Get()
	if err != nil {
		hostname, err = os.Hostname()
		if err != nil {
			return err
		}
		info = map[string]string{"publicHostname": hostname}
	} else {
		hostname = info["publicHostname"]
	}
	log.Println("hostname", hostname)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	key := "/our/machines/" + hostname + "/awsinfo"
	log.Printf("registerMachine %v", key)
	_, err = etcdClient.Set(key, string(data), 600)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	registerMachine()
	ids, err := getContainerIds()
	for _, id := range ids {
		err := registerContainer(id.Id, "start")
		if err != nil {
			log.Panic(err)
		}
	}

	res, err := client.Get("http://ignor.ed/events")
	log.Println(res)
	if err != nil {
		log.Panic(err)
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			log.Panic(err)
		}

		log.Printf("%#v\n", event)
		if event.Status == "start" || event.Status == "stop" {
			registerContainer(event.Id, event.Status)
		}
	}
}
