package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	auth_middleware "github.com/Petr09Mitin/xrust-beze-back/internal/middleware"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	httpparser "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/httpparser"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
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
	authClient    authpb.AuthServiceClient
}

func NewChat(chatService chat_service.ChatService, msgSub *MessageSubscriber, authClient authpb.AuthServiceClient, m *melody.Melody, logger zerolog.Logger, cfg *config.Chat) (*Chat, error) {
	ch := &Chat{
		ChatService:   chatService,
		msgSubscriber: msgSub,
		M:             m,
		logger:        logger,
		cfg:           cfg,
		authClient:    authClient,
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
	chatGroup.Use(auth_middleware.AuthMiddleware(ch.authClient))
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
		custom_errors.WriteHTTPError(c, err)
		return
	}
	ch.logger.Info().Msg("ws connection upgrade successful")
}

// func (ch *Chat) handleNewChatJoin(s *melody.Session) {
// 	userID := strings.TrimSpace(s.Request.URL.Query().Get(userIDQueryParam))
// 	s.Set(UserIDSessionParam, userID)
// }

func (ch *Chat) handleNewChatJoin(s *melody.Session) {
	cookie, err := s.Request.Cookie(auth_middleware.SkillSharingTokenCookieKey)
	if err != nil || cookie.Value == "" {
		ch.logger.Warn().Msg("no auth token in ws connect")
		s.Close()
		return
	}

	resp, err := ch.authClient.ValidateSession(
		s.Request.Context(),
		&authpb.SessionRequest{SessionId: cookie.Value},
	)
	if err != nil || !resp.Valid {
		ch.logger.Warn().Err(err).Msg("invalid auth token in ws")
		s.Close()
		return
	}

	queryUserID := strings.TrimSpace(s.Request.URL.Query().Get(userIDQueryParam))
	if queryUserID != resp.UserId {
		ch.logger.Warn().
			Str("auth_user_id", resp.UserId).
			Str("query_user_id", queryUserID).
			Msg("user_id mismatch detected in handleNewChatJoin - replacing requested user_id with auth_user_id")
	}

	s.Set(UserIDSessionParam, resp.UserId)

	ctx := context.WithValue(s.Request.Context(), "user_id", resp.UserId)
	s.Request = s.Request.WithContext(ctx)
}

func (ch *Chat) handleMessage(ctx context.Context, msg []byte) error {
	parsedMsg := chat_models.Message{}
	err := json.Unmarshal(msg, &parsedMsg)
	if err != nil {
		return err
	}

	authUserID, ok := ctx.Value("user_id").(string)
	if !ok {
		return custom_errors.ErrNoAuthUserID
	}
	// если id авторизованного юзера не соответствует id отправителя, логируем и подменяем на верное
	if parsedMsg.UserID != authUserID {
		ch.logger.Warn().
			Str("auth_user_id", authUserID).
			Str("message_user_id", parsedMsg.UserID).
			Msg("user_id mismatch detected in handleMessage - replacing message user_id with auth_user_id")
		parsedMsg.UserID = authUserID
	}

	ch.logger.Println("msg came to server", parsedMsg)
	switch parsedMsg.Event {
	case chat_models.TextMsgEvent:
		err = ch.ChatService.ProcessTextMessage(ctx, parsedMsg)
	case chat_models.StructurizationEvent:
		err = ch.ChatService.ProcessStructurizationRequest(ctx, parsedMsg)
	case chat_models.VoiceMessageEvent:
		err = ch.ChatService.ProcessVoiceMessage(ctx, parsedMsg)
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
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoChannelID)
		return
	}

	authUserID, ok := auth_middleware.GetUserIDFromGinContext(c)
	if !ok {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoAuthUserID)
		return
	}
	ctx := context.WithValue(c.Request.Context(), "user_id", authUserID)

	limit, offset := httpparser.GetLimitAndOffset(c)
	messages, err := ch.ChatService.GetMessagesByChatID(ctx, channelID, limit, offset)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func (ch *Chat) HandleGetChannelsByUserID(c *gin.Context) {
	userID := strings.TrimSpace(c.Query(userIDQueryParam))
	if userID == "" {
		custom_errors.WriteHTTPError(c, custom_errors.ErrInvalidUserID)
		return
	}

	err := ch.assertUserID(c, userID)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	limit, offset := httpparser.GetLimitAndOffset(c)
	channels, err := ch.ChatService.GetChannelsByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
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
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoUserIDOrPeerID)
		return
	}

	err := ch.assertUserID(c, userID)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	channel, msgs, err := ch.ChatService.GetChannelByUserAndPeerIDs(c.Request.Context(), userID, peerID)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
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
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoMessageID)
		return
	}

	authUserID, ok := auth_middleware.GetUserIDFromGinContext(c)
	if !ok {
		custom_errors.WriteHTTPError(c, custom_errors.ErrNoAuthUserID)
		return
	}
	ctx := context.WithValue(c.Request.Context(), "user_id", authUserID)

	message, err := ch.ChatService.GetMessageByID(ctx, messageID)
	if err != nil {
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func (ch *Chat) assertUserID(c *gin.Context, userID string) error {
	authUserID, ok := auth_middleware.GetUserIDFromGinContext(c)
	if !ok {
		return custom_errors.ErrNoAuthUserID
	}
	if userID != authUserID {
		return custom_errors.ErrAccessDenied
	}
	return nil
}
