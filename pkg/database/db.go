package database

import (
	"github.com/akinbezatoglu/survey-web-app/pkg/database/mongodb"
	"github.com/akinbezatoglu/survey-web-app/pkg/model"
)

// Config represents the configuration of the database interface
type Config struct {
	MongoDB *mongodb.Config
}

// DB is the interface which must be implemented by all db drivers
type DB interface {
	NewConnection(config *Config) (*mongodb.DB, error)
	CloseConnection() error

	CreateUser(u *model.User) error
	GetUser(id string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	SaveUser(u *model.User) error
	DeleteUser(id string) error

	CreateForm(f *model.Form) (string, error)
	GetForm(id string) (*model.Form, error)
	SaveForm(f *model.Form) error
	DeleteForm(id string) error

	CreateQuestion(q *model.Question) (string, error)
	GetQuestion(id string) (*model.Question, error)
	SaveQuestion(q *model.Question) error
	DeleteQuestion(id string) error

	CreateAnswer(q *model.Answer) (string, error)
	GetAnswer(id string) (*model.Answer, error)
	SaveAnswer(q *model.Answer) error
	DeleteAnswer(id string) error
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*mongodb.DB, error) {
	// Use MongoDB
	db, err := mongodb.NewConnection(config.MongoDB)
	if err != nil {
		return nil, err
	}
	return db, nil
}
