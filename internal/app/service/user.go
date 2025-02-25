package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type UserRepository interface {
	CreateUser(ctx context.Context) (string, error)
}

type UserService struct {
	storage   UserRepository
	secretKey string
}

func NewUserService(storage UserRepository, secretKey string) *UserService {
	return &UserService{
		storage: storage,
	}
}

func (s *UserService) Create(ctx context.Context) (string, error) {
	userID, err := s.storage.CreateUser(ctx)
	if err != nil {
		return "", err
	}
	signedUserID, err := s.signUserID(userID)
	if err != nil {
		return "", err
	}
	user := userID + ":" + signedUserID
	return user, nil
}

func (s *UserService) Verify(user string) bool {
	values := strings.Split(user, ":")
	originUserID := values[0]
	signedUserID := values[1]

	expected, err := s.signUserID(originUserID)
	if err != nil {
		return false
	}
	if expected != signedUserID {
		return false
	}
	return true
}

func (s *UserService) signUserID(userID string) (string, error) {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	_, err := h.Write([]byte(userID))
	if err != nil {
		return "", err
	}
	signature := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature), nil
}
