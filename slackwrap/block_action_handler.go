package slackwrap

import (
	"fmt"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/taketsuru-devel/gorilla-microservice-skeleton/skeletonutil"
)

type BlockActionHandlerFunc func(http.ResponseWriter, *http.Request, *slack.Client, *slack.InteractionCallback, *slack.BlockAction) error

type BlockActionHandler interface {
	GetEventHandler() EventHandlerFunc
	GetEventId() string
	GetBlockActionHandler() BlockActionHandlerFunc //on interactive
}

//eventからblockActionIdを取得する関数
type GetEventIdFromEventFunc func(*http.Request, *slack.Client, *slackevents.EventsAPIEvent) string

type blockActionManager struct {
	//eventとinteractive
	handlerMap map[string]BlockActionHandler
	//eventからevent_idを取得する関数
	eventIdFunc GetEventIdFromEventFunc
}

func NewBlockActionManager(f GetEventIdFromEventFunc) *blockActionManager {
	return &blockActionManager{
		handlerMap:  make(map[string]BlockActionHandler),
		eventIdFunc: f,
	}
}

func (b *blockActionManager) Register(h BlockActionHandler) {
	b.handlerMap[h.GetEventId()] = h
}

func (b *blockActionManager) EventHandle() EventHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ev *slackevents.EventsAPIEvent) (interrupt bool, err error) {
		key := b.eventIdFunc(r, cli, ev)
		if key != "" {
			interrupt = true
			if b.handlerMap == nil || len(b.handlerMap) == 0 {
				//ここで処理すべきなのに未定義はエラー扱い
				err = fmt.Errorf("block_action event is not defined")
			} else if h, ok := b.handlerMap[key]; !ok {
				err = fmt.Errorf("eventId:%s is not registered", key)
			} else {
				interrupt, err = h.GetEventHandler()(w, r, cli, ev)
			}
		}
		return
	}
}

func (b *blockActionManager) InteractiveHandle() InteractiveHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, cli *slack.Client, ic *slack.InteractionCallback) (interrupt bool, err error) {

		if ic.BlockActionState != nil && len(ic.BlockActionState.Values) > 0 {
			interrupt = true
			skeletonutil.DebugLog(fmt.Sprintf("block_actions:%#v", ic.BlockActionState), 0)
			if b.handlerMap == nil || len(b.handlerMap) == 0 {
				//ここで処理すべきなのに未定義はエラー扱い
				err = fmt.Errorf("block_action event is not defined")
				return
			}
			//map[string]map[string]BlockAction
			//一次のキーが何を指してるか不明
			//二次のキーはevent_id
			for _, vmap := range ic.BlockActionState.Values {
				for eventId, state := range vmap {
					if h, ok := b.handlerMap[eventId]; ok {
						if err = h.GetBlockActionHandler()(w, r, cli, ic, &state); err != nil {
							return
						}
					} else {
						err = fmt.Errorf("eventId:%s is not registered", eventId)
						return
					}
				}
			}
		}
		return
	}
}
