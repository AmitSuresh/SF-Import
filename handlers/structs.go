package handlers

import (
	"crypto/rsa"

	"go.uber.org/zap"
)

type Handler struct {
	clientKey    string
	clientSecret string
	username     string
	privateKey   *rsa.PrivateKey
	instanceURL  string
	authURL      string
	tokenURL     string
	sobjectsURL  string
	UserAgent    string
	accessToken  string
	refreshToken string
	jwtToken     string
	l            *zap.Logger
}

type FieldMetadata struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type MetadataResponse struct {
	Fields []FieldMetadata `json:"fields"`
}

type FieldAPILabelMapping map[string]string
