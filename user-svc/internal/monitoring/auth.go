package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type HydraClient struct {
	httpClient *http.Client
	adminURL   string
}

func NewHydraClient() *HydraClient {
	return &HydraClient{
		httpClient: &http.Client{},
		adminURL:   "http://127.0.0.1:4445",
	}
}

func AuthInterceptor(hydra *HydraClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md["authorization"]) == 0 {
			return nil, status.Error(codes.Unauthenticated, "No token provided")
		}
		token := md["authorization"][0][7:] // Remove "Bearer " prefix
		data := url.Values{}
		data.Set("token", token)
		reqIntrospect, _ := http.NewRequest("POST", hydra.adminURL+"/admin/oauth2/introspect", bytes.NewBufferString(data.Encode()))
		reqIntrospect.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := hydra.httpClient.Do(reqIntrospect)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "Failed to introspect token")
		}
		defer resp.Body.Close()

		var introspectResp struct {
			Active bool `json:"active"`
		}
		json.NewDecoder(resp.Body).Decode(&introspectResp)
		if !introspectResp.Active {
			return nil, status.Error(codes.Unauthenticated, "Invalid token")
		}
		return handler(ctx, req)
	}
}