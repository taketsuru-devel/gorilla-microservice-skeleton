package slackwrap

import (
	"encoding/json"
	"github.com/slack-go/slack"
	"net/http"
)

//interrupt: これを抜けたら後続のハンドラを飛ばすか errorの場合はinterrupt無視で飛ばす
type InteractiveHandlerFunc func(http.ResponseWriter, *http.Request, *slack.Client, *slack.InteractionCallback) (bool, error)
type InteractiveHandler interface {
	InteractiveHandle() InteractiveHandlerFunc
}

type InteractiveEndpoint struct {
	Client   *slack.Client
	Handlers []InteractiveHandler
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

	for _, h := range i.Handlers {
		if interrupt, err := h.InteractiveHandle()(w, r, i.Client, &payload); err != nil {
			//http error
			w.WriteHeader(http.StatusInternalServerError)
			break
		} else if interrupt {
			break
		}
	}
}
