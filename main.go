package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/taketsuru-devel/gorilla-microservice-skeleton/serverwrap"
	"github.com/taketsuru-devel/gorilla-microservice-skeleton/skeletonutil"
	"github.com/taketsuru-devel/gorilla-microservice-skeleton/slackwrap"
)

const BLOCK_ACTION_IMPL_ID = "sample"

func main() {
	server := serverwrap.NewServer(":13000")

	cli := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	signingSecret := os.Getenv("SIGNING_SECRET")
	f := slackwrap.NewSlackHandlerFactory(cli, &signingSecret, &sampleEventHandler{}, &sampleInteractiveHander{})
	f.InitBlockAction(GetEventIdImpl)
	f.RegisterBlockAction(&blockActionHandlerImpl{})
	server.AddHandle("/events-endpoint", f.CreateEventEndpoint()).Methods("POST")
	server.AddHandle("/interactive", f.CreateInteractiveEndpoint()).Methods("POST")

	server.Start()
	defer server.Stop(60)

	skeletonutil.WaitSignal()
}

type sampleEventHandler struct{}

func (s *sampleEventHandler) EventHandle() slackwrap.EventHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ev *slackevents.EventsAPIEvent) (interrupt bool, err error) {
		interrupt = true
		skeletonutil.InfoLog(fmt.Sprintf("%v", ev.InnerEvent.Data), 0)
		mentionEv, ok := ev.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			err = fmt.Errorf("not mentioned data: %#v", ev.InnerEvent.Data)
			return
		}
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
		return
	}
}

type sampleInteractiveHander struct{}

func (s *sampleInteractiveHander) InteractiveHandle() slackwrap.InteractiveHandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ic *slack.InteractionCallback) (interrupt bool, err error) {
		skeletonutil.InfoLog(fmt.Sprintf("%v", ic.Message.Text), 0)
		return
	})
}

func GetEventIdImpl(r *http.Request, cli *slack.Client, ev *slackevents.EventsAPIEvent) (eventId string) {
	switch innerData := ev.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		//export SLACK_BOT_USERID="<@U****>"
		text := strings.ReplaceAll(innerData.Text, os.Getenv("SLACK_BOT_USERID"), "")
		text = strings.TrimSpace(strings.ReplaceAll(text, "\u00a0", "")) //nbsp
		skeletonutil.InfoLog(text, 0)
		if text == BLOCK_ACTION_IMPL_ID {
			eventId = text
		}
	}
	return
}

type blockActionHandlerImpl struct{}

func (b *blockActionHandlerImpl) GetEventHandler() slackwrap.EventHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ev *slackevents.EventsAPIEvent) (interrupt bool, err error) {
		switch innerData := ev.InnerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			interrupt = true
			//一覧表示してlistにしてイベントを作る
			selectDefault := slack.NewTextBlockObject("plain_text", "未選択", false, false)
			select2 := slack.NewTextBlockObject("plain_text", "選択", false, false)
			select2obj := slack.NewOptionBlockObject("選択", select2, select2)
			opte := slack.NewOptionsSelectBlockElement("static_select", selectDefault, BLOCK_ACTION_IMPL_ID, select2obj)
			notice := slack.NewTextBlockObject("plain_text", "以下から選択してください", false, false)
			mbk := slack.NewSectionBlock(notice, nil, slack.NewAccessory(opte))
			_, _, err = cli.PostMessage(innerData.Channel, slack.MsgOptionBlocks(mbk))
		}

		return
	}
}
func (b *blockActionHandlerImpl) GetEventId() string {
	return BLOCK_ACTION_IMPL_ID
}
func (b *blockActionHandlerImpl) GetBlockActionHandler() slackwrap.BlockActionHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ic *slack.InteractionCallback, b *slack.BlockAction) (err error) {
		channelId := ic.Channel.GroupConversation.Conversation.ID
		skeletonutil.InfoLog(channelId, 0)
		_, _, err = cli.PostMessage(channelId, slack.MsgOptionText(fmt.Sprintf("%sを受け付けました", b.SelectedOption.Value), false))
		return
	}
}
