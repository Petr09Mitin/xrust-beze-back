package user

// Repository определяет интерфейс для работы с хранилищем пользователей
type Repository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(page, limit int) ([]*User, error)
	FindBySkills(skillsToLearn, skillsToShare []string) ([]*User, error)
}

// Service определяет интерфейс для бизнес-логики работы с пользователями
type Service interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(page, limit int) ([]*User, error)
	FindMatchingUsers(userID string) ([]*User, error)
} 