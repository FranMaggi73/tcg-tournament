package handlers

import (
	"net/http"

	"github.com/FranMaggi73/tcg-tournament/backend/internal/models"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/tournament"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FriendshipHandler struct {
	repo *tournament.Repository
}

func NewFriendshipHandler(repo *tournament.Repository) *FriendshipHandler {
	return &FriendshipHandler{
		repo: repo,
	}
}

// AddFriend sends a friend request. (Protected)
func (h *FriendshipHandler) AddFriend(c *gin.Context) {
	var req struct {
		FriendID string `json:"friendId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uidStr := uid.(string)

	if uidStr == req.FriendID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot add yourself as a friend"})
		return
	}

	f := &models.Friendship{
		ID:        uuid.NewString(),
		User1ID:   uidStr,
		User2ID:   req.FriendID,
		Status:    "pending",
	}

	if err := h.repo.CreateFriendship(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, f)
}

// GetFriends returns the list of accepted friends. (Protected)
func (h *FriendshipHandler) GetFriends(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uidStr := uid.(string)

	friends, err := h.repo.GetFriends(c.Request.Context(), uidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if friends == nil {
		friends = []*models.Friendship{}
	}
	c.JSON(http.StatusOK, friends)
}

// UpdateFriendshipStatus accepts or declines a friend request. (Protected)
func (h *FriendshipHandler) UpdateFriendshipStatus(c *gin.Context) {
	friendshipID := c.Param("id")
	var req struct {
		Status string `json:"status"` // "accepted" or "declined"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != "accepted" && req.Status != "declined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'accepted' or 'declined'"})
		return
	}

	if err := h.repo.UpdateFriendshipStatus(c.Request.Context(), friendshipID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friendship status updated"})
}

// GetPendingRequests returns pending friend requests for the current user. (Protected)
func (h *FriendshipHandler) GetPendingRequests(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uidStr := uid.(string)

	requests, err := h.repo.GetPendingRequests(c.Request.Context(), uidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if requests == nil {
		requests = []*models.Friendship{}
	}
	c.JSON(http.StatusOK, requests)
}
