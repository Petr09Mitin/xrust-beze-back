package router

import (
	"context"
	"encoding/json"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/chat"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"net/http"
	"strings"
)

const (
	userIDQueryParam    = "user_id"
	channelIDQueryParam = "channel_id"
)

type Chat struct {
	R             *gin.Engine
	M             *melody.Melody
	msgSubscriber *chat.MessageSubscriber
	ChatService   chat_service.ChatService
}

func NewChat(chatService chat_service.ChatService, msgSub *chat.MessageSubscriber, m *melody.Melody) *Chat {
	ch := &Chat{
		ChatService:   chatService,
		msgSubscriber: msgSub,
		M:             m,
	}
	err := ch.InitWS()
	if err != nil {
		fmt.Println(err)
	}
	ch.InitRouter()
	return ch
}

func (ch *Chat) InitRouter() {
	ch.R = gin.Default()

	chatGroup := ch.R.Group("/chat")
	{
		chatGroup.GET("/ws", ch.HandleWSConn)
		chatGroup.GET("/dialogs")
		chatGroup.GET("/{chatID}}")
	}
}

func (ch *Chat) InitWS() error {
	ch.msgSubscriber.RegisterHandler()
	ch.M.HandleConnect(ch.handleNewChatJoin)

	ch.M.HandleDisconnect(func(s *melody.Session) {
		fmt.Println("dis", s.Request)
	})

	ch.M.HandleMessage(func(s *melody.Session, msg []byte) {
		err := ch.handleTextMessage(s.Request.Context(), msg)
		if err != nil {
			data, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				fmt.Println(err)
				return
			}
			s.Write(data)
		}
	})

	go func() {
		err := ch.msgSubscriber.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (ch *Chat) Start() error {
	fmt.Println("start chat")
	err := ch.R.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

func (ch *Chat) HandleWSConn(c *gin.Context) {
	err := ch.M.HandleRequest(c.Writer, c.Request)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
	}
}

func (ch *Chat) handleNewChatJoin(s *melody.Session) {
	// TODO: add proper auth
	userID := strings.TrimSpace(s.Request.URL.Query().Get(userIDQueryParam))
	channelID := strings.TrimSpace(s.Request.URL.Query().Get(channelIDQueryParam))
	s.Set(chat.UserIDSessionParam, userID)
	s.Set(chat.ChannelIDSessionParam, channelID)
	fmt.Println("conn", s.Request)
}

func (ch *Chat) handleTextMessage(ctx context.Context, msg []byte) error {
	parsedMsg := chat_models.Message{}
	err := json.Unmarshal(msg, &parsedMsg)
	if err != nil {
		return err
	}
	fmt.Println("msg came to server", parsedMsg)
	err = ch.ChatService.ProcessTextMessage(ctx, parsedMsg)
	if err != nil {
		return err
	}

	return nil
}

func (ch *Chat) Stop() error {
	err := ch.msgSubscriber.GracefulStop()
	if err != nil {
		return err
	}

	return nil
}
