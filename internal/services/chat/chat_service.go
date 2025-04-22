package chat_service

import (
	"context"
	"errors"
	"fmt"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/Petr09Mitin/xrust-beze-back/internal/repository/file_client"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	"time"

	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	channelrepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/channel"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/chat"
	structurization_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/structurization"
	user_grpc "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/user"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type ChatService interface {
	ProcessTextMessage(ctx context.Context, message chat_models.Message) error
	ProcessStructurizationRequest(ctx context.Context, message chat_models.Message) error
	ProcessVoiceMessage(ctx context.Context, message chat_models.Message) error
	GetMessagesByChatID(ctx context.Context, chatID string, limit, offset int64) ([]chat_models.Message, error)
	GetChannelsByUserID(ctx context.Context, userID string, limit, offset int64) ([]chat_models.Channel, error)
	GetChannelByUserAndPeerIDs(ctx context.Context, userID, peerID string) (*chat_models.Channel, []chat_models.Message, error)
	GetMessageByID(ctx context.Context, messageID string) (*chat_models.Message, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, in *pb.GetUserByIDRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
}

type ChatServiceImpl struct {
	msgRepo             message_repo.MessageRepo
	channelRepo         channelrepo.ChannelRepository
	fileServiceClient   file_client.FileServiceClient
	structurizationRepo structurization_repo.StructurizationRepository
	userService         UserService
	studyMaterialPub    study_material_repo.StudyMaterialPub
	cfg                 *config.Chat
	logger              zerolog.Logger
}

func NewChatService(
	msgRepo message_repo.MessageRepo,
	channelRepo channelrepo.ChannelRepository,
	fileServiceClient file_client.FileServiceClient,
	structurizationRepo structurization_repo.StructurizationRepository,
	userService UserService,
	studyMaterialPub study_material_repo.StudyMaterialPub,
	logger zerolog.Logger,
	cfg *config.Chat) ChatService {
	return &ChatServiceImpl{
		msgRepo:             msgRepo,
		channelRepo:         channelRepo,
		fileServiceClient:   fileServiceClient,
		structurizationRepo: structurizationRepo,
		userService:         userService,
		studyMaterialPub:    studyMaterialPub,
		cfg:                 cfg,
		logger:              logger,
	}
}

func (c *ChatServiceImpl) ProcessTextMessage(ctx context.Context, msg chat_models.Message) error {
	var newMsg chat_models.Message
	var err error

	switch msg.Type {
	case chat_models.SendMessageType:
		newMsg, err = c.createTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	case chat_models.UpdateMessageType:
		newMsg, err = c.updateTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	case chat_models.DeleteMessageType:
		newMsg, err = c.deleteTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	default:
		return custom_errors.ErrInvalidMessageType
	}

	if err := c.msgRepo.PublishMessage(ctx, newMsg); err != nil {
		c.logger.Err(err)
		return custom_errors.ErrBroadcastingTextMessage
	}
	return nil
}

func (c *ChatServiceImpl) ProcessVoiceMessage(ctx context.Context, msg chat_models.Message) error {
	var newMsg chat_models.Message
	var err error

	switch msg.Type {
	case chat_models.SendMessageType:
		newMsg, err = c.createVoiceMessage(ctx, msg)
		if err != nil {
			return err
		}
	case chat_models.DeleteMessageType:
		newMsg, err = c.deleteVoiceMessage(ctx, msg)
		if err != nil {
			return err
		}
	default:
		return custom_errors.ErrInvalidMessageType
	}

	if err := c.msgRepo.PublishMessage(ctx, newMsg); err != nil {
		c.logger.Err(err)
		return custom_errors.ErrBroadcastingTextMessage
	}
	return nil
}

func (c *ChatServiceImpl) ProcessStructurizationRequest(ctx context.Context, message chat_models.Message) error {
	oldMessage, err := c.msgRepo.GetMessageByID(ctx, message.MessageID)
	if err != nil {
		return err
	}
	channel, err := c.channelRepo.GetChannelByID(ctx, oldMessage.ChannelID)
	if err != nil {
		return err
	}
	oldMessage.SetReceiverIDs(channel.UserIDs)

	prevMessages, err := c.msgRepo.GetPreviousMessagesByMessageCreatedAt(ctx, channel.ID, oldMessage.CreatedAt, 1)
	if err != nil {
		return err
	}
	question := c.concatenateMessages(prevMessages)
	structurized, err := c.trySendStructurizationRequest(ctx, question, oldMessage.Payload)
	if err != nil {
		return err
	}
	newMsg := *oldMessage
	newMsg.Structurized = structurized
	newMsg.UpdatedAt = time.Now().Unix()
	err = c.msgRepo.UpdateMessage(ctx, newMsg)
	if err != nil {
		return err
	}
	newMsg.Type = chat_models.UpdateMessageType
	newMsg.Event = chat_models.StructurizationEvent

	err = c.msgRepo.PublishMessage(ctx, newMsg)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatServiceImpl) createTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	var channel chat_models.Channel
	var err error

	if msg.Payload == "" && len(msg.Attachments) == 0 {
		return chat_models.Message{}, custom_errors.ErrInvalidMessage
	}

	if msg.ChannelID == "" {
		if msg.UserID == "" || msg.PeerID == "" {
			return chat_models.Message{}, custom_errors.ErrInvalidMessage
		}
		channel, err = c.channelRepo.GetByUserIDs(ctx, []string{msg.UserID, msg.PeerID})
		if err != nil {
			if errors.Is(err, custom_errors.ErrNotFound) {
				created := time.Now().Unix()
				channel, err = c.channelRepo.InsertChannel(ctx, chat_models.Channel{
					UserIDs: []string{
						msg.UserID,
						msg.PeerID,
					},
					Created: created,
					Updated: created,
				})
				if err != nil {
					return msg, err
				}
			} else {
				return msg, err
			}
		}
	} else {
		channel, err = c.channelRepo.GetChannelByID(ctx, msg.ChannelID)
		if err != nil {
			return msg, err
		}
	}

	createdAt := time.Now().Unix()
	if len(msg.Attachments) > 0 {
		msg.Attachments, err = c.fileServiceClient.MoveTempFilesToAttachments(ctx, msg.Attachments)
		if err != nil {
			return chat_models.Message{}, err
		}
		// publish potential materials for studymateriald to process
		// if we failed - log and continue
		prevMsgs, err := c.msgRepo.GetPreviousMessagesByMessageCreatedAt(ctx, channel.ID, createdAt, 10)
		if err != nil {
			c.logger.Error().Err(err).Str("channel_id", channel.ID).Msg("unable to get previous messages in studymateriald sending")
		} else {
			err = c.publishAttachmentsToProcess(ctx, &msg, prevMsgs)
			if err != nil {
				c.logger.Error().Err(err).Any("msg", msg).Msg("unable to publish for studymateriald in create msg")
			}
		}
	} else {
		msg.Attachments = make([]string, 0)
	}

	newMsg := chat_models.Message{
		Event:       chat_models.TextMsgEvent,
		Type:        msg.Type,
		ChannelID:   channel.ID,
		UserID:      msg.UserID,
		PeerID:      msg.PeerID,
		Payload:     msg.Payload,
		Attachments: msg.Attachments,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	newMsg.SetReceiverIDs(channel.UserIDs)
	newMsg, err = c.msgRepo.InsertMessage(ctx, newMsg)
	if err != nil {
		c.logger.Err(err)
		return msg, custom_errors.ErrBroadcastingTextMessage
	}
	c.logger.Printf("new message saved: %+v\n", newMsg)
	return newMsg, nil
}

func (c *ChatServiceImpl) createVoiceMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	var channel chat_models.Channel
	var err error
	if msg.Voice == "" {
		c.logger.Error().Msg("empty voice msg in create")
		return chat_models.Message{}, custom_errors.ErrInvalidMessage
	}

	if msg.ChannelID == "" {
		if msg.UserID == "" || msg.PeerID == "" {
			return chat_models.Message{}, custom_errors.ErrInvalidMessage
		}
		channel, err = c.channelRepo.GetByUserIDs(ctx, []string{msg.UserID, msg.PeerID})
		if err != nil {
			if errors.Is(err, custom_errors.ErrNotFound) {
				created := time.Now().Unix()
				channel, err = c.channelRepo.InsertChannel(ctx, chat_models.Channel{
					UserIDs: []string{
						msg.UserID,
						msg.PeerID,
					},
					Created: created,
					Updated: created,
				})
				if err != nil {
					return msg, err
				}
			} else {
				return msg, err
			}
		}
	} else {
		channel, err = c.channelRepo.GetChannelByID(ctx, msg.ChannelID)
		if err != nil {
			return msg, err
		}
	}

	filename, err := c.fileServiceClient.MoveTempFileToVoiceMessages(ctx, msg.Voice)
	if err != nil {
		return chat_models.Message{}, err
	}

	createdAt := time.Now().Unix()
	newMsg := chat_models.Message{
		Event:         chat_models.VoiceMessageEvent,
		Type:          msg.Type,
		ChannelID:     channel.ID,
		UserID:        msg.UserID,
		PeerID:        msg.PeerID,
		Voice:         filename,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
		VoiceDuration: msg.VoiceDuration,
		Payload:       "",
	}
	newMsg.SetReceiverIDs(channel.UserIDs)
	newMsg, err = c.msgRepo.InsertMessage(ctx, newMsg)
	if err != nil {
		c.logger.Err(err)
		return msg, custom_errors.ErrBroadcastingTextMessage
	}
	c.logger.Printf("new message saved: %+v\n", newMsg)
	return newMsg, nil
}

func (c *ChatServiceImpl) updateTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	var channel chat_models.Channel
	var err error

	if msg.MessageID == "" {
		return msg, custom_errors.ErrNoMessageID
	}
	if msg.Payload == "" && len(msg.Attachments) == 0 {
		return chat_models.Message{}, custom_errors.ErrInvalidMessage
	}
	if msg.ChannelID == "" {
		return msg, custom_errors.ErrNoChannelID
	}
	channel, err = c.channelRepo.GetChannelByID(ctx, msg.ChannelID)
	if err != nil {
		return msg, err
	}
	oldMsg, err := c.msgRepo.GetMessageByID(ctx, msg.MessageID)
	if err != nil {
		return msg, err
	}

	newAttachmentsMap := make(map[string]any, len(msg.Attachments))
	for _, attachment := range msg.Attachments {
		newAttachmentsMap[attachment] = struct{}{}
	}
	attachmentsToDelete := make([]string, 0)
	attachmentsToPreserve := make([]string, 0, len(oldMsg.Attachments))
	oldAttachmentsMap := make(map[string]any, len(oldMsg.Attachments))
	for _, attachment := range oldMsg.Attachments {
		oldAttachmentsMap[attachment] = struct{}{}
		// если старого аттача нет в новых - удаляем
		if _, ok := newAttachmentsMap[attachment]; !ok {
			attachmentsToDelete = append(attachmentsToDelete, attachment)
		} else {
			attachmentsToPreserve = append(attachmentsToPreserve, attachment)
		}
	}
	if len(attachmentsToDelete) > 0 {
		err = c.fileServiceClient.DeleteAttachments(ctx, attachmentsToDelete)
		if err != nil {
			return chat_models.Message{}, err
		}
	}
	attachmentsToCreate := make([]string, 0, len(oldMsg.Attachments))
	for _, attachment := range msg.Attachments {
		// если нового аттача нет в старых - создаем
		if _, ok := oldAttachmentsMap[attachment]; !ok {
			attachmentsToCreate = append(attachmentsToCreate, attachment)
		}
	}
	var filenames []string
	if len(attachmentsToCreate) > 0 {
		filenames, err = c.fileServiceClient.MoveTempFilesToAttachments(ctx, attachmentsToCreate)
		if err != nil {
			return chat_models.Message{}, err
		}
		// publish potential materials for studymateriald to process
		// if we failed - log and continue
		prevMsgs, err := c.msgRepo.GetPreviousMessagesByMessageCreatedAt(ctx, channel.ID, oldMsg.CreatedAt, 10)
		if err != nil {
			c.logger.Error().Err(err).Str("channel_id", channel.ID).Msg("unable to get previous messages in studymateriald sending")
		} else {
			err = c.publishAttachmentsToProcess(ctx, &msg, prevMsgs)
			if err != nil {
				c.logger.Error().Err(err).Any("msg", msg).Msg("unable to publish for studymateriald in update msg")
			}
		}
	}

	updatedAt := time.Now().Unix()
	newMsg := chat_models.Message{
		MessageID:   oldMsg.MessageID,
		Event:       msg.Event,
		Type:        msg.Type,
		ChannelID:   channel.ID,
		UserID:      oldMsg.UserID,
		Payload:     msg.Payload,
		CreatedAt:   oldMsg.CreatedAt,
		UpdatedAt:   updatedAt,
		Attachments: append(attachmentsToPreserve, filenames...),
	}
	err = c.msgRepo.UpdateMessage(ctx, newMsg)
	if err != nil {
		return msg, custom_errors.ErrBroadcastingTextMessage
	}
	newMsg.SetReceiverIDs(channel.UserIDs)
	c.logger.Printf("message updated: %+v\n", newMsg)
	return newMsg, nil
}

func (c *ChatServiceImpl) deleteTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	var channel chat_models.Channel
	var err error

	if msg.MessageID == "" {
		return msg, custom_errors.ErrNoMessageID
	}
	oldMsg, err := c.msgRepo.GetMessageByID(ctx, msg.MessageID)
	if err != nil {
		return msg, err
	}

	if msg.ChannelID == "" {
		return msg, custom_errors.ErrNoChannelID
	}
	channel, err = c.channelRepo.GetChannelByID(ctx, oldMsg.ChannelID)
	if err != nil {
		return msg, err
	}

	if len(oldMsg.Attachments) > 0 {
		err = c.fileServiceClient.DeleteAttachments(ctx, oldMsg.Attachments)
		if err != nil {
			return msg, err
		}
	}

	err = c.msgRepo.DeleteMessage(ctx, *oldMsg)
	if err != nil {
		return msg, custom_errors.ErrBroadcastingTextMessage
	}

	c.logger.Printf("message deleted: %+v\n", msg)
	msg.SetReceiverIDs(channel.UserIDs)
	return msg, nil
}

func (c *ChatServiceImpl) deleteVoiceMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	var channel chat_models.Channel
	var err error

	if msg.MessageID == "" {
		return msg, custom_errors.ErrNoMessageID
	}

	oldMsg, err := c.msgRepo.GetMessageByID(ctx, msg.MessageID)
	if err != nil {
		return msg, err
	}

	if oldMsg.ChannelID == "" {
		return msg, custom_errors.ErrNoChannelID
	}
	channel, err = c.channelRepo.GetChannelByID(ctx, oldMsg.ChannelID)
	if err != nil {
		return msg, err
	}

	err = c.fileServiceClient.DeleteVoiceMessage(ctx, oldMsg.Voice)
	if err != nil {
		return msg, err
	}

	err = c.msgRepo.DeleteMessage(ctx, *oldMsg)
	if err != nil {
		return msg, custom_errors.ErrBroadcastingTextMessage
	}

	c.logger.Printf("message deleted: %+v\n", oldMsg)
	msg.SetReceiverIDs(channel.UserIDs)
	return msg, nil
}

func (c *ChatServiceImpl) GetMessagesByChatID(ctx context.Context, chatID string, limit, offset int64) ([]chat_models.Message, error) {
	if limit == 0 || limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return c.msgRepo.GetMessagesByChannelID(ctx, chatID, limit, offset)
}

func (c *ChatServiceImpl) GetChannelsByUserID(ctx context.Context, userID string, limit, offset int64) ([]chat_models.Channel, error) {
	channels, err := c.channelRepo.GetChannelsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	for i, channel := range channels {
		msgs, err := c.msgRepo.GetMessagesByChannelID(ctx, channel.ID, 1, 0)
		if err != nil {
			c.logger.Error().Err(err).Msg(fmt.Sprintf("error getting messages by channel %s", channel.ID))
			continue
		}
		if len(msgs) > 0 {
			channels[i].LastMessage = &msgs[0]
		}

		for _, userID := range channel.UserIDs {
			res, err := c.userService.GetUserByID(ctx, &pb.GetUserByIDRequest{
				Id: userID,
			})
			if err != nil {
				c.logger.Error().Err(err).Msg(fmt.Sprintf("unable to get user %s", userID))
				continue
			}
			user, err := user_grpc.ConvertProtoToDomain(res.GetUser())
			if err != nil {
				c.logger.Error().Err(err).Msg(fmt.Sprintf("unable to convert to domain user %s", userID))
				continue
			}
			channels[i].Users = append(channels[i].Users, *user)
		}
	}

	return channels, nil
}

func (c *ChatServiceImpl) concatenateMessages(messages []chat_models.Message) string {
	res := ""
	for _, msg := range messages {
		res += msg.Payload + " \n"
	}
	return res
}

func (c *ChatServiceImpl) trySendStructurizationRequest(ctx context.Context, question, answer string) (string, error) {
	newCtx, cancel := context.WithTimeout(
		ctx,
		time.Duration(c.cfg.Services.StructurizationService.Timeout)*time.Second,
	)
	defer cancel()
	i := c.cfg.Services.StructurizationService.MaxRetries
loop:
	for i > 0 {
		select {
		case <-newCtx.Done():
			return "", custom_errors.ErrRequestTimeout
		default:
			i--
			structurized, err := c.structurizationRepo.SendStructRequest(newCtx, question, answer)
			if err != nil {
				c.logger.Error().Err(err).Msg(fmt.Sprintf("trySendStructurizationRequest failed, %d retries remaining", i))
				continue loop
			}
			return structurized.Explanation, nil
		}
	}

	return "", custom_errors.ErrMaxRetriesExceeded
}

func (c *ChatServiceImpl) GetChannelByUserAndPeerIDs(ctx context.Context, userID, peerID string) (*chat_models.Channel, []chat_models.Message, error) {
	channel, err := c.channelRepo.GetByUserIDs(ctx, []string{userID, peerID})
	if err != nil {
		return nil, nil, err
	}
	for _, userID := range channel.UserIDs {
		res, err := c.userService.GetUserByID(ctx, &pb.GetUserByIDRequest{
			Id: userID,
		})
		if err != nil {
			c.logger.Error().Err(err).Msg(fmt.Sprintf("unable to get user %s", userID))
			continue
		}
		user, err := user_grpc.ConvertProtoToDomain(res.GetUser())
		if err != nil {
			c.logger.Error().Err(err).Msg(fmt.Sprintf("unable to convert to domain user %s", userID))
			continue
		}
		channel.Users = append(channel.Users, *user)
	}
	msgs, err := c.msgRepo.GetMessagesByChannelID(ctx, channel.ID, 200, 0)
	if err != nil {
		return nil, nil, err
	}
	return &channel, msgs, nil
}

func (c *ChatServiceImpl) GetMessageByID(ctx context.Context, messageID string) (*chat_models.Message, error) {
	return c.msgRepo.GetMessageByID(ctx, messageID)
}

func (c *ChatServiceImpl) publishAttachmentsToProcess(ctx context.Context, msg *chat_models.Message, prevMsgs []chat_models.Message) error {
	prevMsgsTexts := make([]string, 0, len(prevMsgs))
	for _, prevMsg := range prevMsgs {
		prevMsgsTexts = append(prevMsgsTexts, prevMsg.Payload)
	}
	for _, attachment := range msg.Attachments {
		err := c.studyMaterialPub.PublishAttachmentToParse(ctx, &study_material_models.AttachmentToParse{
			Filename:         attachment,
			AuthorID:         msg.UserID,
			CurrMessageText:  msg.Payload,
			PrevMessageTexts: prevMsgsTexts,
		})
		if err != nil {
			c.logger.Error().Err(err).Msg(fmt.Sprintf("unable to publish attachment to process: %s", attachment))
			return err
		}
	}

	return nil
}
