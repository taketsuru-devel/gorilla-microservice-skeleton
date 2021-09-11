package slackwrap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type EventSubscribeHandlerFunc func(http.ResponseWriter, *http.Request, *slackevents.EventsAPIEvent)
type EventSubscribeHandler interface {
	Handle() EventSubscribeHandlerFunc
}

type EventSubscribeEndpoint struct {
	Handler       EventSubscribeHandler
	SigningSecret *string
}

func (h *EventSubscribeEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//前処理
	//bodyの取得
	/*
		//GetBody()はserver側では使えないみたい
		bodyReader, err := r.GetBody()
		if err != nil {
			return nil, err
		}
	*/
	bodyReader := r.Body
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		slackHandlerErrorResponse(w, err, 0)
		return
	}
	defer bodyReader.Close()

	//署名の検証
	if err := verifySlackSecret(r.Header, h.SigningSecret, &body); err != nil {
		slackHandlerErrorResponse(w, err, 0)
		return
	}

	//event型の取得
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		slackHandlerErrorResponse(w, err, 0)
		return
	}

	//slackからのAPI検証リクエストならここで処理する
	if done, err := handleAPIVerificationRequest(&eventsAPIEvent, &body, w); err != nil {
		slackHandlerErrorResponse(w, err, 0)
		return
	} else if !done {
		//API検証リクエストじゃなかったので処理続行
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			h.Handler.Handle()(w, r, &eventsAPIEvent)
		} else {
			slackHandlerErrorResponse(w, fmt.Errorf("undefined event:%s", eventsAPIEvent.Type), 0)
		}
	}
}

func verifySlackSecret(header http.Header, signingSecret *string, body *[]byte) error {
	sv, err := slack.NewSecretsVerifier(header, *signingSecret)
	if err != nil {
		return err
	}
	if _, err := sv.Write(*body); err != nil {
		return err
	}
	if err := sv.Ensure(); err != nil {
		return err
	}
	return nil
}

func handleAPIVerificationRequest(ev *slackevents.EventsAPIEvent, body *[]byte, w http.ResponseWriter) (bool, error) {
	//slackからのAPI検証
	if ev.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal(*body, &r)
		if err != nil {
			return false, err
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		return true, nil
	}
	return false, nil
}
