package gapi

import (
	"context"
	"fmt"
	"github.com/mariobasic/simplebank/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authorizationHeader = "Authorization"
	authorizationBearer = "Bearer"
)

func (s *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader) // expecting two // <authorization-type> <authorization-data> // Bearer xxx
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid authorization header")
	}
	authType := fields[0]
	if authType != authorizationBearer {
		return nil, fmt.Errorf("invalid authorization type: %s", authType)
	}
	accessToken := fields[1]
	payload, err := s.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}