package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

func getContainer(client http.Client, id string) ([]byte, error) {
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
	fmt.Println(proto, addr)
	return net.Dial("unix", "/var/run/docker.sock")
}

func getContainerIds(c http.Client) ([]ContainerId, error) {
	res, err := c.Get("http://ignor.ed/containers/json")
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
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

func registerContainer(client http.Client, id string, event string) error {
	data, err := getContainer(client, id)

	var container Container
	fmt.Printf("%#v\n", string(data))
	err = json.Unmarshal(data, &container)
	if err != nil {
		return err
	} else {
		fmt.Printf("%#v\n", container)
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
	fmt.Printf("registerInEtcd %v, %v", id, data)
	return nil
}

func deregisterFromEtcd(id string) error {
	fmt.Printf("deregisterFromEtcd %v", id)
	return nil
}

func main() {
	tr := &http.Transport{
		Dial: fakeDial,
	}
	client := http.Client{Transport: tr}

	ids, err := getContainerIds(client)
	fmt.Printf("%#v %v", ids, err)
	for _, id := range ids {
		err := registerContainer(client, id.Id, "start")
		if err != nil {
			log.Fatal(err)
		}
	}

	res, err := client.Get("http://ignor.ed/events")
	fmt.Println(res)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		fmt.Printf("%#v\n", event)
		if event.Status == "start" || event.Status == "stop" {
			registerContainer(client, event.Id, event.Status)
		}
	}
}
