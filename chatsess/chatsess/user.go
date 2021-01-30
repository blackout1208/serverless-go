package chatsess

import (
	"errors"
	"fmt"
	"html"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type User struct {
	Username string
	Password string
}

func NewUser(name, pwd string) *User {
	return &User{
		Username: html.EscapeString(name),
		Password: NewPassword(pwd),
	}
}

func (u *User) Put(sess *session.Session) error {
	dbc := dynamodb.New(sess)

	_, err := dbc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("ch_users"),
		Item: map[string]*dynamodb.AttributeValue{
			"Username": {S: aws.String(u.Username)},
			"Password": {S: aws.String(u.Password)},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func GetDBUser(uname string, sess *session.Session) (*User, error) {
	dbc := dynamodb.New(sess)

	uname = html.EscapeString(uname)

	dbres, err := dbc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("ch_user"),
		Key: map[string]*dynamodb.AttributeValue{
			"Username": {S: aws.String(uname)},
		},
	})
	if err != nil {
		return nil, err
	}

	if dbres.Item == nil {
		return nil, errors.New("No user found")
	}

	pwv, ok := dbres.Item["Password"]
	if !ok {
		return nil, fmt.Errorf("User has no password: %s", uname)
	}
	return &User{Username: uname, Password: *(pwv.S)}, nil
}

func GetDBUserPass(uname, pwd string, sess *session.Session) (*User, error) {
	u, err := GetDBUser(uname, sess)
	if err != nil {
		return nil, err
	}

	ok := CheckPassword(pwd, u.Password)
	if !ok {
		return nil, fmt.Errorf("Username or Password incorrect")
	}

	return u, nil
}
