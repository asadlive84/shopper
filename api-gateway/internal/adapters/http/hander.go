package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	pb "github.com/asadlive84/shopper-proto/golang/user"
	"github.com/asadlive84/shopper/api-gateway/internal/adapters/grpc"
	"github.com/asadlive84/shopper/api-gateway/internal/event"
	"github.com/asadlive84/shopper/api-gateway/internal/ports"
	"github.com/asadlive84/shopper/api-gateway/internal/rabbitmq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
)

type HydraClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	adminURL   string
	publicURL  string
}

type Handler struct {
	apiService    ports.APIPort
	jwtSecret     string
	logger        *zap.Logger
	rabbitClient  *rabbitmq.Client
	hydra         *HydraClient
	websocketConn *websocket.Conn
	eventManager  *event.EventManager
}

func NewHydraClient(logger *zap.Logger) *HydraClient {
	return &HydraClient{
		logger:     logger,
		httpClient: &http.Client{},
		adminURL:   "http://127.0.0.1:4445",
		publicURL:  "http://127.0.0.1:4444",
	}
}

func NewHandler(apiService ports.APIPort, jwtSecret string, logger *zap.Logger, rabbitClient *rabbitmq.Client, grpcClient *grpc.UserGRPCClient) *Handler {
	return &Handler{
		apiService:   apiService,
		jwtSecret:    jwtSecret,
		logger:       logger,
		rabbitClient: rabbitClient,
		hydra:        NewHydraClient(logger),
		eventManager: event.NewEventManager(logger, rabbitClient, grpcClient),
	}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	r.Use(LoggerMiddleware(h.logger), TracingMiddleware(), PrometheusMiddleware())
	r.Use(CORSMiddleware())

	r.POST("/login", h.Login)
	// r.GET("/callback", h.Callback)

	r.POST("/user", h.CreateUser)

	protected := r.Group("/", JWTAuthMiddleware(h.jwtSecret, h.logger))
	protected.GET("/user/:id", h.GetUser)

	protected.GET("/users", h.ListUsers)
	protected.PUT("/user/:id", h.UpdateUser)

	r.GET("/ws", CORSMiddleware(), h.WebSocketHandler)

}

func (h *Handler) Login(c *gin.Context) {
	var req pb.AuthenticateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.apiService.AuthenticateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Warn("Authentication failed", zap.String("email", req.GetEmail()), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	msg := "User logged in: " + req.GetEmail()
	if err := h.rabbitClient.Publish(c.Request.Context(), "user_exchange", msg); err != nil {
		h.logger.Error("Failed to publish to RabbitMQ", zap.Error(err))
	}
	c.JSON(http.StatusOK, gin.H{"token": res.GetToken(), "user": res.GetUser()})
}

func (h *Handler) Callback(c *gin.Context) {
	code := c.Query("code")
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://127.0.0.1:8080/callback")
	data.Set("client_id", "your-client-id")
	data.Set("client_secret", "your-client-secret")

	req, _ := http.NewRequest("POST", h.hydra.publicURL+"/oauth2/token", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.hydra.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to get token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	c.SetCookie("access_token", tokenResp.AccessToken, 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/user")
}

func (h *Handler) AuthMiddleware(c *gin.Context) {
	token, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	data := url.Values{}
	data.Set("token", token)
	req, _ := http.NewRequest("POST", h.hydra.adminURL+"/admin/oauth2/introspect", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.hydra.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to introspect token", zap.Error(err))
		c.Redirect(http.StatusFound, "/login")
		return
	}
	defer resp.Body.Close()

	var introspectResp struct {
		Active bool `json:"active"`
	}
	json.NewDecoder(resp.Body).Decode(&introspectResp)
	if !introspectResp.Active {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.Next()
}

func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.apiService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req pb.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.apiService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.String("email", req.GetEmail()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) ListUsers(c *gin.Context) {
	var req pb.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to parse list users query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.apiService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req pb.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = id
	user, err := h.apiService.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
