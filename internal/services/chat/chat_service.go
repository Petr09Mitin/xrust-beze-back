package chat_service

import chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"

type MessageRepository interface {
	Create(ctx context.Context, message chat_models.Message) error
	FindAll(ctx context.Context) ([]chat_models.Message, error)
	FindByID(ctx context.Context, id string) (chat_models.Message, error)
}

type ChatService interface {
	ProcessTextMessage(message chat_models.Message) error
	GetAllMessages(ctx context.Context) ([]chat_models.Message, error)
	GetMessageByID(ctx context.Context, id string) (chat_models.Message, error)
}

type ChatServiceImpl struct {
	repo MessageRepository
}

func NewChatService(repo MessageRepository) ChatService {
	return &ChatServiceImpl{
		repo: repo,
	}
}

func (c *ChatServiceImpl) ProcessTextMessage(message chat_models.Message) error {
	ctx := context.Background()
	err := c.repo.Create(ctx, message)
	if err != nil {
		return err
	}
	
	fmt.Println("Сообщение сохранено в MongoDB:", message.Content)
	return nil
}

func (c *ChatServiceImpl) GetAllMessages(ctx context.Context) ([]chat_models.Message, error) {
	return c.repo.FindAll(ctx)
}

func (c *ChatServiceImpl) GetMessageByID(ctx context.Context, id string) (chat_models.Message, error) {
	return c.repo.FindByID(ctx, id)
}
