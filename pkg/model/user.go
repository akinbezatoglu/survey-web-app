package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

//	var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// User is the structure of a user
type User struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name,omitempty"`
	Lastname string             `json:"lastname" bson:"lastname,omitempty"`
	Email    string             `json:"email" bson:"email,omitempty"`
	Password string             `json:"password" bson:"password,omitempty"`
	Forms    []string           `bson:"forms,omitempty"`
	Token    string             `json:"token" bson:"token,omitempty"`
}

type ViewUser struct {
	ID       string   `json:"_id" bson:"_id,omitempty"`
	Name     string   `json:"name" bson:"name,omitempty"`
	Lastname string   `json:"lastname" bson:"lastname,omitempty"`
	Forms    []string `bson:"forms,omitempty"`
	Email    string   `json:"email" bson:"email,omitempty"`
	Password string   `json:"password" bson:"password,omitempty"`
	Token    string   `json:"token" bson:"token,omitempty"`
}

// HashPassword hashed the password of the user
func (u *User) HashPassword() error {
	key, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(key)

	return nil
}

// MatchPassword returns true if the hashed user password matches the password
func (u *User) MatchPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	return err == nil
}
