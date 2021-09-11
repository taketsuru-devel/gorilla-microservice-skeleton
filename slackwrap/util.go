package slackwrap

import (
	"net/http"

	"github.com/followedwind/gorilla-microservice-skeleton/util"
)

func slackHandlerErrorResponse(w http.ResponseWriter, err error, addStack int) {
	util.ErrorLog(err.Error(), 1+addStack)
	w.WriteHeader(http.StatusInternalServerError)
}
