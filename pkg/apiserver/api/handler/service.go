package handler

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/fest-research/IoT-apiserver/pkg/apiserver/proxy"
)

type KubeService struct {
	proxy proxy.IServerProxy
}

func NewKubeService(proxy proxy.IServerProxy) KubeService {
	return KubeService{proxy: proxy}
}

func (this KubeService) Register(ws *restful.WebService) {
	// List services
	ws.Route(
		ws.Method("GET").
			Path("/services").
			To(this.listServices).
			Returns(http.StatusOK, "OK", nil).
			Writes(nil),
	)

	// Watch services
	ws.Route(
		ws.Method("GET").
			Path("/watch/services").
			To(this.watchServices).
			Returns(http.StatusOK, "OK", nil).
			Writes(nil),
	)
}

func (this KubeService) listServices(req *restful.Request, resp *restful.Response) {
	response, err := this.proxy.Get(req)
	if err != nil {
		handleInternalServerError(resp, err)
	}

	resp.AddHeader("Content-Type", "application/json")
	resp.Write(response)
}

func (this KubeService) watchServices(req *restful.Request, resp *restful.Response) {
	response, err := this.proxy.Get(req)
	if err != nil {
		handleInternalServerError(resp, err)
	}

	resp.AddHeader("Content-Type", "application/json")
	resp.Write(response)
}