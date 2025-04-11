package chat

import (
	"context"
	"encoding/json"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	httpparser "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/httpparser"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
	"net/http"
	"strings"
)

const (
	userIDQueryParam = "user_id"
	peerIDQueryParam = "peer_id"
)

type Chat struct {
	R             *gin.Engine
	M             *melody.Melody
	msgSubscriber *MessageSubscriber
	ChatService   chat_service.ChatService
	logger        zerolog.Logger
	cfg           *config.Chat
}

func NewChat(chatService chat_service.ChatService, msgSub *MessageSubscriber, m *melody.Melody, logger zerolog.Logger, cfg *config.Chat) (*Chat, error) {
	ch := &Chat{
		ChatService:   chatService,
		msgSubscriber: msgSub,
		M:             m,
		logger:        logger,
		cfg:           cfg,
	}
	err := ch.InitWS()
	if err != nil {
		ch.logger.Error().Err(err).Msg("init websocket error")
		return nil, err
	}
	ch.InitRouter()
	return ch, nil
}

func (ch *Chat) InitRouter() {
	ch.R = gin.Default()
	ch.R.Use(middleware.CORSMiddleware())

	chatGroup := ch.R.Group("/api/v1/chat")
	{
		chatGroup.GET("/ws", ch.HandleWSConn)
		chatGroup.GET("/:channelID", ch.HandleGetMessagesByChannelID)
		chatGroup.GET("/channels/by-peer", ch.handleGetChannelByUserAndPeerIDs)
		chatGroup.GET("/channels", ch.HandleGetChannelsByUserID)
		chatGroup.GET("/messages/:messageID", ch.GetMessagebyID)
	}
}

func (ch *Chat) InitWS() error {
	ch.msgSubscriber.RegisterHandler()
	ch.M.HandleConnect(ch.handleNewChatJoin)

	ch.M.HandleDisconnect(func(s *melody.Session) {
		ch.logger.Println("dis", s.Request)
	})

	ch.M.HandleMessage(func(s *melody.Session, msg []byte) {
		err := ch.handleMessage(s.Request.Context(), msg)
		if err != nil {
			data, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				ch.logger.Error().Err(err).Msg("unable to marshal error")
				return
			}
			err = s.Write(data)
			if err != nil {
				ch.logger.Error().Err(err).Msg("unable to write error to response")
				return
			}
			return
		}
	})

	go func() {
		err := ch.msgSubscriber.Run()
		if err != nil {
			ch.logger.Err(err)
			return
		}
	}()

	return nil
}

func (ch *Chat) Start() error {
	ch.logger.Println("start chat http")
	err := ch.R.Run(fmt.Sprintf(":%d", ch.cfg.HTTP.Port))
	if err != nil {
		return err
	}

	return nil
}

func (ch *Chat) HandleWSConn(c *gin.Context) {
	err := ch.M.HandleRequest(c.Writer, c.Request)
	if err != nil {
		ch.logger.Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
		return
	}
}

func (ch *Chat) handleNewChatJoin(s *melody.Session) {
	userID := strings.TrimSpace(s.Request.URL.Query().Get(userIDQueryParam))
	s.Set(UserIDSessionParam, userID)
}

func (ch *Chat) handleMessage(ctx context.Context, msg []byte) error {
	parsedMsg := chat_models.Message{}
	err := json.Unmarshal(msg, &parsedMsg)
	if err != nil {

		return err
	}
	ch.logger.Println("msg came to server", parsedMsg)
	switch parsedMsg.Event {
	case chat_models.TextMsgEvent:
		err = ch.ChatService.ProcessTextMessage(ctx, parsedMsg)
	case chat_models.StructurizationEvent:
		err = ch.ChatService.ProcessStructurizationRequest(ctx, parsedMsg)
	default:
		err = custom_errors.ErrInvalidMessageEvent
	}
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

func (ch *Chat) HandleGetMessagesByChannelID(c *gin.Context) {
	channelID := strings.TrimSpace(c.Param("channelID"))
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid channelID",
		})
		return
	}

	limit, offset := httpparser.GetLimitAndOffset(c)
	messages, err := ch.ChatService.GetMessagesByChatID(c.Request.Context(), channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func (ch *Chat) HandleGetChannelsByUserID(c *gin.Context) {
	userID := strings.TrimSpace(c.Query(userIDQueryParam))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userID",
		})
		return
	}
	limit, offset := httpparser.GetLimitAndOffset(c)
	channels, err := ch.ChatService.GetChannelsByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
	})
}

func (ch *Chat) handleGetChannelByUserAndPeerIDs(c *gin.Context) {
	userID := strings.TrimSpace(c.Query(userIDQueryParam))
	peerID := strings.TrimSpace(c.Query(peerIDQueryParam))
	if userID == "" || peerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userID or peerID",
		})
		return
	}

	channel, msgs, err := ch.ChatService.GetChannelByUserAndPeerIDs(c.Request.Context(), userID, peerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userID or peerID",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"channel":  channel,
		"messages": msgs,
	})
}

func (ch *Chat) GetMessagebyID(c *gin.Context) {
	messageID := strings.TrimSpace(c.Param("messageID"))
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid messageID",
		})
		return
	}

	message, err := ch.ChatService.GetMessageByID(c.Request.Context(), messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}
