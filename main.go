package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/taketsuru-devel/gorilla-microservice-skeleton/serverwrap"
	"github.com/taketsuru-devel/gorilla-microservice-skeleton/slackwrap"
	"github.com/taketsuru-devel/gorilla-microservice-skeleton/util"
)

func main() {
	server := serverwrap.NewServer(":13000")

	server.AddHandle("/events-endpoint", sampleEventSubscribeEndpoint()).Methods("POST")
	server.AddHandle("/interactive", sampleInteractiveEndpoint()).Methods("POST")

	server.Start()
	defer server.Stop(60)

	util.WaitSignal()
}

type sampleSubscribeHander struct{}

func (s *sampleSubscribeHander) Handle() slackwrap.EventSubscribeHandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request, ev *slackevents.EventsAPIEvent) {
		util.InfoLog(fmt.Sprintf("%v", ev.InnerEvent.Data), 0)
		mentionEv, ok := ev.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			return
		}
		cli := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
		headerText := slack.NewTextBlockObject("mrkdwn", "*interactive test*", false, false)
		headerSection := slack.NewSectionBlock(headerText, nil, nil)

		testBtnTxt := slack.NewTextBlockObject("plain_text", "test message", false, false)
		testBtn := slack.NewButtonBlockElement("", "testCommand", testBtnTxt)
		actionBlock := slack.NewActionBlock("", testBtn)
		msg := slack.MsgOptionBlocks(
			headerSection,
			actionBlock,
		)

		cli.PostMessage(mentionEv.Channel, msg)
	})
}

func sampleEventSubscribeEndpoint() *slackwrap.EventSubscribeEndpoint {
	signingSecret := os.Getenv("SIGNING_SECRET")
	return &slackwrap.EventSubscribeEndpoint{
		Handler:       &sampleSubscribeHander{},
		SigningSecret: &signingSecret,
	}
}

type sampleInteractiveHander struct{}

func (s *sampleInteractiveHander) Handle() slackwrap.InteractiveHandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request, ic *slack.InteractionCallback) {
		util.InfoLog(fmt.Sprintf("%v", ic.Message.Text), 0)
	})
}

func sampleInteractiveEndpoint() *slackwrap.InteractiveEndpoint {
	return &slackwrap.InteractiveEndpoint{
		Handler: &sampleInteractiveHander{},
	}
}
