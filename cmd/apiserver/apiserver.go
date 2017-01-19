package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/fest-research/IoT-apiserver/pkg/apiserver/api"
	"github.com/fest-research/IoT-apiserver/pkg/apiserver/api/handler"
	"github.com/fest-research/IoT-apiserver/pkg/apiserver/proxy"
	kube "github.com/fest-research/IoT-apiserver/pkg/kubernetes"
	"github.com/spf13/pflag"
)

var (
	argApiserverHost = pflag.String("apiserver", "", "Kubernetes api server address")
	argPort          = pflag.Int("port", 8083, "Port to listen on")
	argKubeconfig    = pflag.String("kubeconfig", "./kubeconfig.yaml", "absolute path to the kubeconfig file")
)

const (
	version  = "v1"
	rootPath = "/api/" + version
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	log.Printf("Using HTTP port: %d", *argPort)
	if *argApiserverHost == "" {

		log.Fatal("Parameter 'apiserver' not defined. Please define kubernetes apiserver address.")
	}

	if *argKubeconfig == "" {
		log.Fatal("Parameter 'kubeconfig' not defined." +
			" Please provide a 'kubeconfig' file to access the kubernetes apiserver.")
	}

	// Create a client for the kubernetes apis
	clientset := kube.NewClientset(*argKubeconfig)

	// Create api installer
	installer := api.APIInstaller{Root: rootPath, Version: version}

	// Create api proxy TODO: poll server and check if address is correct
	proxy := proxy.NewServerProxy(*argApiserverHost)

	// Create service factory
	serviceFactory := handler.NewServiceFactory(clientset, proxy)

	ws := installer.NewWebService()
	installer.Install(ws, serviceFactory.GetRegisteredServices())

	restful.Add(ws)
	http.ListenAndServe(fmt.Sprintf(":%d", *argPort), nil)
}