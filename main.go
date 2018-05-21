package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
)

var (
	name   = flag.String("name", "echo-01", "")
	addr   = flag.String("address", ":8181", "echo server address")
	consul = flag.String("consul-url", "http://localhost:8500", "consul server url")
)

func main() {
	flag.Parse()

	fmt.Printf(`%s listen and serve at %s`, *name, *addr)

	// consul client
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	FatalIfError(err)

	//
	host, port := HostPort(*addr)

	// register
	// https://www.consul.io/api/catalog.html#register-entity
	reg := &api.CatalogRegistration{
		Node:           *name,
		SkipNodeUpdate: true,
		Service: &api.AgentService{
			ID:      *name,
			Service: *name,
			Address: host,
			Port:    port,
		},
	}
	_, err = client.Catalog().Register(reg, nil)
	FatalIfError(err)

	// server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(formattedRequest(r))
	})
	err = http.ListenAndServe(*addr, nil)
	FatalIfError(err)
}

func formattedRequest(r *http.Request) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v %v %v\n", r.Method, r.URL, r.Proto))
	buffer.WriteString(fmt.Sprintf("Host: %v", r.Host))

	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			buffer.WriteString(fmt.Sprintf("%v: %v", name, h))
		}
	}

	// TODO: handle content tpye
	body, _ := ioutil.ReadAll(r.Body)
	buffer.WriteString(fmt.Sprintf("\n\n"))
	buffer.Write(body)

	return buffer.Bytes()
}

func HostPort(address string) (host string, port int) {
	chunks := strings.Split(address, ":")
	if len(chunks) > 1 {
		host = chunks[0]
		port, _ = strconv.Atoi(chunks[1])
		return
	}

	host = address
	return
}

func FatalIfError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}

}
