package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"cdek_wishlist/internal/domain"
	"cdek_wishlist/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine  *gin.Engine
	service *usecase.Service
	tokens  usecase.TokenManager
}

func New(service *usecase.Service, tokens usecase.TokenManager) *Server {
	gin.SetMode(gin.ReleaseMode)
	server := &Server{
		engine:  gin.New(),
		service: service,
		tokens:  tokens,
	}
	server.engine.Use(gin.Logger(), gin.Recovery())
	server.registerRoutes()
	return server
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) registerRoutes() {
	s.engine.POST("/auth/register", s.register)
	s.engine.POST("/auth/login", s.login)

	auth := s.engine.Group("/api", s.authMiddleware())
	{
		auth.POST("/wishlists", s.createWishlist)
		auth.GET("/wishlists", s.listWishlists)
		auth.GET("/wishlists/:id", s.getWishlist)
		auth.PUT("/wishlists/:id", s.updateWishlist)
		auth.DELETE("/wishlists/:id", s.deleteWishlist)

		auth.POST("/wishlists/:id/items", s.createItem)
		auth.PUT("/wishlists/:id/items/:item_id", s.updateItem)
		auth.DELETE("/wishlists/:id/items/:item_id", s.deleteItem)
	}

	s.engine.GET("/public/wishlists/:token", s.getPublicWishlist)
	s.engine.POST("/public/wishlists/:token/items/:item_id/book", s.bookItem)
}

type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type wishlistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
}

type itemRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Degree      int    `json:"degree"`
}

func (s *Server) register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	user, err := s.service.Register(c.Request.Context(), usecase.RegisterInput{Email: req.Email, Password: req.Password})
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

func (s *Server) login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	token, err := s.service.Login(c.Request.Context(), usecase.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) createWishlist(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	var req wishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	eventDate, err := parseOptionalTime(req.EventDate)
	if err != nil {
		badRequest(c, err)
		return
	}
	wishlist, err := s.service.CreateWishlist(c.Request.Context(), userID, usecase.WishlistInput{Name: req.Name, Description: req.Description, EventDate: eventDate})
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	writeWishlist(c, http.StatusCreated, wishlist)
}

func (s *Server) listWishlists(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlists, err := s.service.ListWishlists(c.Request.Context(), userID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	response := make([]gin.H, 0, len(wishlists))
	for i := range wishlists {
		response = append(response, wishlistToJSON(&wishlists[i], false))
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) getWishlist(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	wishlist, err := s.service.GetWishlist(c.Request.Context(), userID, wishlistID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	writeWishlist(c, http.StatusOK, wishlist)
}

func (s *Server) updateWishlist(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	var req wishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	eventDate, err := parseOptionalTime(req.EventDate)
	if err != nil {
		badRequest(c, err)
		return
	}
	wishlist, err := s.service.UpdateWishlist(c.Request.Context(), userID, wishlistID, usecase.WishlistInput{Name: req.Name, Description: req.Description, EventDate: eventDate})
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	writeWishlist(c, http.StatusOK, wishlist)
}

func (s *Server) deleteWishlist(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	if err := s.service.DeleteWishlist(c.Request.Context(), userID, wishlistID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) createItem(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	item, err := s.service.CreateItem(c.Request.Context(), userID, wishlistID, usecase.ItemInput(req))
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, itemToJSON(item))
}

func (s *Server) updateItem(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	itemID, err := parseID(c.Param("item_id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	item, err := s.service.UpdateItem(c.Request.Context(), userID, wishlistID, itemID, usecase.ItemInput(req))
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, itemToJSON(item))
}

func (s *Server) deleteItem(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		unauthorized(c)
		return
	}
	wishlistID, err := parseID(c.Param("id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	itemID, err := parseID(c.Param("item_id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	if err := s.service.DeleteItem(c.Request.Context(), userID, wishlistID, itemID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) getPublicWishlist(c *gin.Context) {
	wishlist, err := s.service.GetPublicWishlist(c.Request.Context(), c.Param("token"))
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	writeWishlist(c, http.StatusOK, wishlist)
}

func (s *Server) bookItem(c *gin.Context) {
	itemID, err := parseID(c.Param("item_id"))
	if err != nil {
		badRequest(c, err)
		return
	}
	if err := s.service.BookItem(c.Request.Context(), c.Param("token"), itemID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" || len(authorization) < 8 || authorization[:7] != "Bearer " {
			unauthorized(c)
			c.Abort()
			return
		}
		userID, err := s.tokens.ParseUserToken(authorization[7:])
		if err != nil {
			unauthorized(c)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func userIDFromContext(c *gin.Context) (int64, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func parseOptionalTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func writeWishlist(c *gin.Context, status int, wishlist *domain.Wishlist) {
	items := make([]gin.H, 0, len(wishlist.Items))
	for i := range wishlist.Items {
		items = append(items, itemToJSON(&wishlist.Items[i]))
	}
	c.JSON(status, wishlistToJSON(wishlist, true, items...))
}

func wishlistToJSON(wishlist *domain.Wishlist, includeToken bool, items ...gin.H) gin.H {
	response := gin.H{
		"id":          wishlist.ID,
		"user_id":     wishlist.UserID,
		"name":        wishlist.Name,
		"description": wishlist.Description,
		"event_date":  wishlist.EventDate,
		"created_at":  wishlist.CreatedAt,
	}
	if includeToken {
		response["token"] = wishlist.Token
	}
	if items != nil {
		response["items"] = items
	}
	return response
}

func itemToJSON(item *domain.Item) gin.H {
	return gin.H{
		"id":            item.ID,
		"wishlist_id":   item.WishlistID,
		"name":          item.Name,
		"description":   item.Description,
		"link":          item.Link,
		"desire_degree": item.DesireDegree,
		"booked":        item.Booked,
	}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}

func writeUsecaseError(c *gin.Context, err error) {
	switch err {
	case domain.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrConflict, domain.ErrAlreadyBooked:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case usecase.ErrUnauthorized:
		unauthorized(c)
	case usecase.ErrInvalidInput:
		badRequest(c, err)
	case usecase.ErrEmailExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
