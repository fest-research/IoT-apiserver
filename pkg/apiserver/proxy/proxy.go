package proxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/api"
)

type IServerProxy interface {
	Put(*restful.Request, v1.APIResource) ([]byte, error)
	Get(*restful.Request, v1.APIResource) ([]byte, error)
	List(*restful.Request, v1.APIResourceList) ([]byte, error)
	Post(*restful.Request, v1.APIResource) ([]byte, error)
	Patch(*restful.Request, v1.APIResource) ([]byte, error)
	Watch(*v1.APIResource, *api.ListOptions) (watch.Interface, error)
}

type ServerProxy struct {
	kubeClient    *dynamic.Client
	serverAddress string
}

func NewServerProxy(kubeClient *dynamic.Client, address string) ServerProxy {
	// TODO: remove the serverAddress when kubeClient is used everywhere
	return ServerProxy{kubeClient: kubeClient, serverAddress: address}
}

func (this ServerProxy) List(req *restful.Request, resource v1.APIResourceList) ([]byte, error) {
	requestPath := this.serverAddress + this.removePathParams(req.Request.URL)
	log.Printf("[Proxy] LIST Request (%s)", requestPath)

	// TODO: replace this with a call to the kube-client
	r, err := http.Get(requestPath)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Proxy] LIST Response (%s): %s", requestPath, string(body))
	return body, nil
}

func (this ServerProxy) Get(req *restful.Request, resource v1.APIResource) ([]byte, error) {
	requestPath := this.serverAddress + this.removePathParams(req.Request.URL)
	log.Printf("[Proxy] GET Request (%s)", requestPath)

	// TODO: replace this with a call to the kube-client
	r, err := http.Get(requestPath)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Proxy] GET Response (%s): %s", requestPath, string(body))
	return body, nil
}

func (this ServerProxy) Put(req *restful.Request, resource v1.APIResource) ([]byte, error) {
	requestPath := this.serverAddress + this.removePathParams(req.Request.URL)

	defer req.Request.Body.Close()
	reqBody, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("[Proxy] PUT Request (%s): %s", requestPath, string(reqBody))

	// TODO: replace this with a call to the kube-client
	r, err := http.NewRequest("PUT", requestPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Proxy] PUT Response (%s): %s", requestPath, string(body))
	return body, nil
}

func (this ServerProxy) Post(req *restful.Request, resource v1.APIResource) ([]byte, error) {
	requestPath := this.serverAddress + this.removePathParams(req.Request.URL)

	defer req.Request.Body.Close()
	reqBody, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("[Proxy] POST Request (%s): %s", requestPath, string(reqBody))

	// TODO: replace this with a call to the kube-client
	r, err := http.Post(requestPath, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Proxy] POST Response (%s): %s", requestPath, string(body))
	return body, nil
}

func (this ServerProxy) Patch(req *restful.Request, resource v1.APIResource) ([]byte, error) {
	requestPath := this.serverAddress + this.removePathParams(req.Request.URL)

	defer req.Request.Body.Close()
	reqBody, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("[Proxy] PATCH Request (%s): %s", requestPath, string(reqBody))

	r, err := http.NewRequest("PATCH", requestPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[Proxy] PATCH Response (%s): %s", requestPath, string(body))
	return body, nil
}

func (this ServerProxy) Watch(resource *v1.APIResource, listOptions *api.ListOptions) (
	watch.Interface, error) {
	watcher, err := this.kubeClient.
		Resource(resource, api.NamespaceAll).
		Watch(listOptions)

	return watcher, err
}

// Remove everything after '?' in url path (FOR TESTS ONLY!)
func (this ServerProxy) removePathParams(url *url.URL) string {
	path := url.String()
	if strings.Contains(path, "?") {
		path = path[:strings.Index(path, "?")]
	}
	return path
}
