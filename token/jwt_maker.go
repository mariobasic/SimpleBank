package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"time"
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("secret key too short, less than %d", minSecretKeySize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

func (j *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	return token, payload, err
}

func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			log.Println("That's not even a token", err)
			return nil, ErrInvalidToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			log.Println("Invalid signature", err)
			return nil, ErrInvalidToken
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			log.Println("Token is either expired or not active yet", err)
			return nil, ErrInvalidToken
		case errors.Is(err, ErrInvalidToken):
			log.Println("Invalid token", err)
			return nil, ErrInvalidToken
		default:
			log.Println("Couldn't handle this token:", err)
			return nil, ErrInvalidToken
		}
	}

	jwtPayload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return jwtPayload, nil

}
