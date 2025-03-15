package router

import (
	"encoding/json"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"net/http"
)

type Chat struct {
	R           *gin.Engine
	M           *melody.Melody
	ChatService chat_service.ChatService

	mongoRepo   *mongodb.MessageRepository
}

func (chat *Chat) InitRouter() {
	chat.R = gin.Default()
	chat.R.GET("/ws", chat.HandleWSConn)

	messageGroup := chat.R.Group("/messages")
	{
		messageGroup.GET("", func(c *gin.Context) {
			messages, err := chat.ChatService.GetAllMessages(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, messages)
		})

		messageGroup.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			message, err := chat.ChatService.GetMessageByID(c, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, message)
		})
	}
}

func (chat *Chat) InitWS() {
	chat.M = melody.New()
	chat.M.HandleConnect(func(s *melody.Session) {
		s.Keys["user_id"] = s.Request.URL.Query().Get("user_id")
		fmt.Println(s.Request)
	})

	chat.M.HandleDisconnect(func(s *melody.Session) {
		fmt.Println("dis", s.Request)
	})

	chat.M.HandleMessage(func(s *melody.Session, msg []byte) {
		err := chat.handleTextMessage(msg)
		if err != nil {
			err := custom_errors.NewCustomError(err.Error())
			xdd, _ := err.MarshalJSON()
			s.Write(xdd)
		}
	})
}

func NewChat(chatService chat_service.ChatService) *Chat {
	mongoConfig := mongodb.Config{
		URI:        "mongo_db:27017",
		Database:   "xrust_beze",
		Username:   "admin",
		Password:   "admin",
		AuthSource: "admin",
	}

	client, err := mongodb.NewConnection(mongoConfig)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}

	messageRepo := mongodb.NewMessageRepository(client, "xrust_beze", "chats")

	chat := &Chat{
		ChatService: chatService,
		mongoRepo:   messageRepo,
	}
	chat.InitWS()
	chat.InitRouter()
	return chat
}

func (chat *Chat) Start() error {
	fmt.Println("start chat")
	err := chat.R.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

func (chat *Chat) HandleWSConn(c *gin.Context) {
	err := chat.M.HandleRequest(c.Writer, c.Request)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
	}
}

func (chat *Chat) handleTextMessage(msg []byte) error {
	parsedMsg := chat_models.Message{}
	err := json.Unmarshal(msg, &parsedMsg)
	if err != nil {
		return err
	}

	err = chat.ChatService.ProcessTextMessage(parsedMsg)
	if err != nil {
		return err
	}

	return nil
}
