package slackwrap

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type EventHandlerFunc func(http.ResponseWriter, *http.Request, *slack.Client, *slackevents.EventsAPIEvent) (bool, error)
type EventHandler interface {
	EventHandle() EventHandlerFunc
}

type EventEndpoint struct {
	Client        *slack.Client
	Handlers      []EventHandler
	SigningSecret *string
}

func (ee *EventEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if err := verifySlackSecret(r.Header, ee.SigningSecret, &body); err != nil {
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
		//bodyが必要なのでhandlers配列で処理できない
		slackHandlerErrorResponse(w, err, 0)
		return
	} else if !done {
		for _, h := range ee.Handlers {
			if interrupt, err := h.EventHandle()(w, r, ee.Client, &eventsAPIEvent); err != nil {
				//http error
				w.WriteHeader(http.StatusInternalServerError)
				break
			} else if interrupt {
				break
			}
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
