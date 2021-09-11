package serverwrap

import (
	"context"
	"errors"
	"github.com/followedwind/gorilla-microservice-skeleton/util"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type ServerWrap struct {
	server *http.Server
	mux    *mux.Router
}

func NewServer(addr string) *ServerWrap {
	sw := ServerWrap{}
	sw.mux = mux.NewRouter()
	sw.server = &http.Server{
		Addr:         addr,
		Handler:      sw.mux,
		ReadTimeout:  10 * time.Second, //request読み込み
		WriteTimeout: 20 * time.Second, //response出力
		IdleTimeout:  10 * time.Second, //再利用(headerにkeep-aliveが含まれる場合)
	}

	return &sw
}

func (sw *ServerWrap) AddHandle(path string, handler http.Handler) *mux.Route {
	return sw.mux.Handle(path, handler)
}

func (sw *ServerWrap) Start() {
	go func() {
		util.InfoLog("Server listening", 0)
		if serverErr := sw.server.ListenAndServe(); !errors.Is(serverErr, http.ErrServerClosed) {
			util.ErrorLog(serverErr.Error(), 0)
		}
	}()
}

func (sw *ServerWrap) Stop(timeoutSecond int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	defer cancel()
	if err := sw.server.Shutdown(ctx); err != nil {
		util.ErrorLog(err.Error(), 0)
	} else {
		util.InfoLog("Server Done", 0)
	}
}
