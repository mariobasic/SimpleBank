package gapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	// http
	grpcGatewayUserAgent = "grpcgateway-user-agent"
	xForwardedForHeader  = "x-forwarded-for"
	// grpc
	userAgent = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (s *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if usrAgents := md.Get(grpcGatewayUserAgent); len(usrAgents) > 0 {
			mtdt.UserAgent = usrAgents[0]
		}
		if usrAgents := md.Get(userAgent); len(usrAgents) > 0 {
			mtdt.UserAgent = usrAgents[0]
		}

		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}
