package services

import "server/internal/models"

type UserService interface {
	Create(user *models.User) (*models.User, error)
	CreateByGoogle(username string, email string, googleID string, pictureURL string) (*models.User, error)
	CreateByRegistration(username string, hashedPassword string) (*models.User, error)

	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)

	SetUserPrivilage(userID uint64, privileage models.Privileage)

	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type userService struct {
}

func (model *UserModel) GetOrRegisterGoogle(username string, email string, googleID string, pictureURL string) (*User, error)

func (model *UserModel) CreateUserGoogle(username string, email string, googleID string, pictureURL string) (*User, error)

func (model *UserModel) CreateUserRegistration(username string, hashedPassword string) (*User, error)

func (model *UserModel) CreateUser(user *User) (*User, error)

func (model *UserModel) UsernameExists(username string) (bool, error)

func (model *UserModel) GetUserById(id uint64) (*User, error)

func (model *UserModel) GetUserByUsername(name string) (*User, error)

func (model *UserModel) GetUserByEmail(email string) (*User, error)

func (model *UserModel) SetUserPrivilage(userID uint64, privileage Privileage) error
