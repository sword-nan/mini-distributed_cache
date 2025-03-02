package server

import (
	"distributed_cache/common"
	"distributed_cache/service"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var DefaultServiceName = "/_Cache/"

type HTTPPool struct {
	self     string
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: DefaultServiceName,
	}
}

func (h *HTTPPool) log(format string, v ...any) {
	if common.DEBUG {
		log.Printf(format, v...)
	}
}

func (h *HTTPPool) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// TODO: generate error message based on the error
	if !strings.HasPrefix(req.URL.Path, h.basePath) {
		msg := fmt.Sprintf("HTTPPool server unexpected path: %s", req.URL.Path)
		h.log("server-%s [ERROR]: HTTPPool server unexpected path: %s", h.self, req.URL.Path)
		http.Error(resp, msg, http.StatusBadRequest)
		return
	}
	// basePath/groupName/key required
	parttens := strings.SplitN(req.URL.Path[len(h.basePath):], "/", 2)
	if len(parttens) != 2 {
		h.log("server-%s [ERROR]: bad request: %s", h.self, req.URL.Path)
		http.Error(resp, "bad request: "+req.URL.Path, http.StatusBadRequest)
		return
	}
	serviceName, key := parttens[0], parttens[1]
	service, err := service.GetService(serviceName)
	if err != nil {
		h.log("server-%s [ERROR]: %s", h.self, err.Error())
		http.Error(resp, "no such service: "+serviceName, http.StatusNotFound)
		return
	}
	h.log("server-%s [GET]: service[%s] key[%s]", h.self, serviceName, key)
	value, err := service.Get(key)
	if err != nil {
		h.log("server-%s [ERROR]: %s", h.self, err.Error())
		http.Error(resp, "key not found", http.StatusInternalServerError)
		return
	}
	resp.Write(value)
}
