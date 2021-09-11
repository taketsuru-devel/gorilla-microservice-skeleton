package slackwrap

import (
	"encoding/json"
	"github.com/slack-go/slack"
	"net/http"
)

type InteractiveHandlerFunc func(http.ResponseWriter, *http.Request, *slack.InteractionCallback)
type InteractiveHandler interface {
	Handle() InteractiveHandlerFunc
}

type InteractiveEndpoint struct {
	Handler InteractiveHandler
}

func (i *InteractiveEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//情報取得
	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		slackHandlerErrorResponse(w, err, 0)
		//ここでエラーだとChannelの取得もできない
		return
	}

	i.Handler.Handle()(w, r, &payload)
}
