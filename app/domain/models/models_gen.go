// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/99designs/gqlgen/graphql"
)

type CreateUserInput struct {
	FirstName    *string         `json:"firstName,omitempty"`
	LastName     *string         `json:"lastName,omitempty"`
	Email        *string         `json:"email,omitempty"`
	Username     string          `json:"username"`
	Password     string          `json:"password"`
	ProfileImage *graphql.Upload `json:"profileImage,omitempty"`
}

type CreateUserPayload struct {
	FirstName       *string `json:"firstName,omitempty"`
	LastName        *string `json:"lastName,omitempty"`
	Email           *string `json:"email,omitempty"`
	Username        string  `json:"username"`
	ProfileImageURL *string `json:"profileImageURL,omitempty"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginPayload struct {
	AccessToken      string `json:"accessToken"`
	IDToken          string `json:"idToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	NotBeforePolicy  int    `json:"notBeforePolicy"`
	SessionState     string `json:"sessionState"`
	Scope            string `json:"scope"`
}

type Mutation struct {
}

type Query struct {
}
