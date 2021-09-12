package slackwrap

import (
	"net/http"

	"github.com/taketsuru-devel/gorilla-microservice-skeleton/skeletonutil"
)

func slackHandlerErrorResponse(w http.ResponseWriter, err error, addStack int) {
	skeletonutil.ErrorLog(err.Error(), 1+addStack)
	w.WriteHeader(http.StatusInternalServerError)
}
