package handlers

import (
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/FranMaggi73/tcg-tournament/backend/internal/models"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/tournament"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TournamentHandler struct {
	repo  *tournament.Repository
	swiss *tournament.SwissService
}

func NewTournamentHandler(repo *tournament.Repository, swiss *tournament.SwissService) *TournamentHandler {
	return &TournamentHandler{
		repo:  repo,
		swiss: swiss,
	}
}

// CreateTournament creates a new tournament. (Protected)
func (h *TournamentHandler) CreateTournament(c *gin.Context) {
	var t models.Tournament
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	// Generate tournament ID and invite code
	t.ID = uuid.NewString()
	t.CreatedBy = uidStr
	t.InviteCode = uuid.NewString()[:8]
	t.Date = time.Now()

	if t.Status == "" {
		t.Status = "registration"
	}
	if t.CurrentRound == 0 && t.TotalRounds == 0 {
		t.TotalRounds = 0
	}

	if t.Format != "BO1" && t.Format != "BO3" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Must be BO1 or BO3"})
		return
	}

	if err := h.repo.CreateTournament(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// GetTournament returns a tournament's details. (Public)
func (h *TournamentHandler) GetTournament(c *gin.Context) {
	id := c.Param("id")
	t, err := h.repo.GetTournament(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// JoinTournamentByCode allows players to join using a code. (Public/Auth)
func (h *TournamentHandler) JoinTournamentByCode(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.repo.GetTournamentByInviteCode(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired invite code"})
		return
	}

	if t.Status != "registration" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registration is closed for this tournament"})
		return
	}

	// Use email from request if provided, otherwise try to get it from auth token
	email := req.Email
	if email == "" {
		uid, exists := c.Get("uid")
		if exists {
			uidStr := uid.(string)
			// In a real app, we'd fetch the user profile from Firestore to get the email
			// For now, we'll use a placeholder or require it if not found
			email = "user_" + uidStr + "@example.com"
		}
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required for registration"})
		return
	}

	exists, err := h.repo.PlayerExists(c.Request.Context(), t.ID, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking player existence"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player with this email is already registered"})
		return
	}

	p := &models.Player{
		ID:     uuid.NewString(),
		Name:   req.Name,
		Email:  email,
		Status: "active",
	}

	if err := h.repo.CreatePlayer(c.Request.Context(), t.ID, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined tournament", "tournamentId": t.ID, "player": p})
}

// RegisterPlayer is kept for legacy/direct API use but redirects to internal logic
func (h *TournamentHandler) RegisterPlayer(c *gin.Context) {
	tournamentID := c.Param("id")
	var p models.Player
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.Status != "registration" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registration is closed"})
		return
	}

	exists, err := h.repo.PlayerExists(c.Request.Context(), tournamentID, p.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking player existence"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player already registered"})
		return
	}

	if err := h.repo.CreatePlayer(c.Request.Context(), tournamentID, &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// GetStandings returns the current leaderboard. (Public)
func (h *TournamentHandler) GetStandings(c *gin.Context) {
	tournamentID := c.Param("id")
	players, err := h.repo.GetPlayersByTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if players == nil {
		players = []*models.Player{}
	}
	c.JSON(http.StatusOK, players)
}

// NextRound triggers the pairing logic for the next round. (Judge Only)
func (h *TournamentHandler) NextRound(c *gin.Context) {
	tournamentID := c.Param("id")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can generate pairings"})
		return
	}

	// Change status to playing when first round pairings are generated
	if t.Status == "registration" {
		t.Status = "playing"
		if err := h.repo.UpdateTournament(c.Request.Context(), t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tournament status"})
			return
		}
	}

	matches, err := h.swiss.GeneratePairings(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pairings generated and persisted to Firestore", "matches": matches})
}

// UpdateMatchResult records the winner of a match and updates standings. (Judge Only)
func (h *TournamentHandler) UpdateMatchResult(c *gin.Context) {
	tournamentID := c.Param("id")
	matchID := c.Param("matchId")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can update results"})
		return
	}

	var scoreReq struct {
		Player1Score int `json:"player1Score"`
		Player2Score int `json:"player2Score"`
		RoundID      string `json:"roundId"`
	}
	if err := c.ShouldBindJSON(&scoreReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch t.Format {
	case "BO3":
		// Valid BO3 scores: 2-0, 2-1, 0-2, 1-2 (victory) OR 1-1 (draw)
		isVictory := (scoreReq.Player1Score == 2 && (scoreReq.Player2Score == 0 || scoreReq.Player2Score == 1)) ||
			(scoreReq.Player2Score == 2 && (scoreReq.Player1Score == 0 || scoreReq.Player1Score == 1))
		isDraw := scoreReq.Player1Score == 1 && scoreReq.Player2Score == 1
		if !(isVictory || isDraw) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BO3 score. Valid scores are: 2-0, 2-1, 0-2, 1-2, or 1-1 for draw"})
			return
		}
	case "BO1":
		if !((scoreReq.Player1Score == 1 && scoreReq.Player2Score == 0) || (scoreReq.Player2Score == 1 && scoreReq.Player1Score == 0)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BO1 score. One player must have 1 win and the other 0"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported tournament format"})
		return
	}

	// Fetch the existing match to get player IDs and derive the winner
	existingMatch, err := h.repo.GetMatch(c.Request.Context(), tournamentID, scoreReq.RoundID, matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Build the updated match preserving player IDs from the existing match
	m := &models.Match{
		ID:           existingMatch.ID,
		RoundID:      existingMatch.RoundID,
		Player1ID:    existingMatch.Player1ID,
		Player2ID:    existingMatch.Player2ID,
		Player1Score: scoreReq.Player1Score,
		Player2Score: scoreReq.Player2Score,
		Status:       "completed",
	}

	// Derive winner from scores: empty WinnerID means draw
	switch {
	case scoreReq.Player1Score > scoreReq.Player2Score:
		m.WinnerID = m.Player1ID
	case scoreReq.Player2Score > scoreReq.Player1Score:
		m.WinnerID = m.Player2ID
	default:
		m.WinnerID = "" // draw
	}

	if err := h.swiss.ProcessMatchResult(c.Request.Context(), tournamentID, m.RoundID, matchID, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match result atomically"})
		return
	}

	if err := h.swiss.UpdateStandings(c.Request.Context(), tournamentID); err != nil {
		log.Printf("error updating standings: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match result updated and standings recalculated"})
}

// RemovePlayer completely deletes a player from a tournament. Only allowed during registration. (Judge Only)
func (h *TournamentHandler) RemovePlayer(c *gin.Context) {
	tournamentID := c.Param("id")
	playerID := c.Param("playerId")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can remove players"})
		return
	}

	if t.Status != "registration" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Players can only be removed during registration. Use drop instead."})
		return
	}

	if err := h.repo.DeletePlayer(c.Request.Context(), tournamentID, playerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player removed from tournament"})
}

// UpdatePlayerStatus allows the judge to drop a player. (Judge Only)
func (h *TournamentHandler) UpdatePlayerStatus(c *gin.Context) {
	tournamentID := c.Param("id")
	playerID := c.Param("playerId")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can update player status"})
		return
	}

	var statusReq struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdatePlayerStatus(c.Request.Context(), tournamentID, playerID, statusReq.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player status updated successfully"})
}

// RollbackRound removes the current round and all subsequent rounds. (Judge Only)
func (h *TournamentHandler) RollbackRound(c *gin.Context) {
	tournamentID := c.Param("id")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can rollback rounds"})
		return
	}

	if t.CurrentRound == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot rollback: no rounds have been generated yet"})
		return
	}

	// 1. Delete the current round and all its matches from Firestore
	rounds, err := h.repo.GetRoundsByTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}

	for _, rnd := range rounds {
		if rnd.RoundNumber == t.CurrentRound {
			// Delete all matches in this round first
			matches, err := h.repo.GetMatchesByRound(c.Request.Context(), tournamentID, rnd.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch matches for round"})
				return
			}
			for _, m := range matches {
				if err := h.repo.DeleteMatch(c.Request.Context(), tournamentID, rnd.ID, m.ID); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete match"})
					return
				}
			}
			// Then delete the round document itself
			if err := h.repo.DeleteRound(c.Request.Context(), tournamentID, rnd.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete round"})
				return
			}
		}
	}

	// 2. Reset all player stats and recalculate from remaining rounds
	players, err := h.repo.GetPlayersByTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}

	for _, p := range players {
		p.TotalScore = 0
		p.Wins = 0
		p.Losses = 0
		p.Draws = 0
		p.OMW = 0
		p.GW = 0
		p.OGW = 0
		h.repo.UpdatePlayer(c.Request.Context(), tournamentID, p)
	}

	// 3. Decrement the round counter
	t.CurrentRound--

	// If we rolled back to round 0, the tournament goes back to registration
	if t.CurrentRound == 0 {
		t.Status = "registration"
	}

	if err := h.repo.UpdateTournament(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tournament round"})
		return
	}

	// 4. Recalculate standings from remaining matches (if any rounds left)
	if t.CurrentRound > 0 {
		if err := h.swiss.UpdateStandings(c.Request.Context(), tournamentID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate standings"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Round rolled back, matches deleted, and standings recalculated"})
}

// ExportStandings returns a formatted summary of the tournament standings. (Public)
func (h *TournamentHandler) ExportStandings(c *gin.Context) {
	tournamentID := c.Param("id")
	players, err := h.repo.GetPlayersByTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings for export"})
		return
	}

	sort.Slice(players, func(i, j int) bool {
		if players[i].TotalScore != players[j].TotalScore {
			return players[i].TotalScore > players[j].TotalScore
		}
		if players[i].OMW != players[j].OMW {
			return players[i].OMW > players[j].OMW
		}
		if players[i].GW != players[j].GW {
			return players[i].GW > players[j].GW
		}
		return players[i].OGW > players[j].OGW
	})

	type ExportPlayer struct {
		Rank       int     `json:"rank"`
		Name       string  `json:"name"`
		TotalScore int     `json:"totalScore"`
		OMW        float64 `json:"omw"`
		GW         float64 `json:"gw"`
		OGW        float64 `json:"ogw"`
	}

	var exportData []ExportPlayer
	for i, p := range players {
		exportData = append(exportData, ExportPlayer{
			Rank:       i + 1,
			Name:       p.Name,
			TotalScore: p.TotalScore,
			OMW:        p.OMW,
			GW:         p.GW,
			OGW:        p.OGW,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tournamentID": tournamentID,
		"standings":    exportData,
	})
}

// DeleteTournament deletes a tournament if it's in registration or completed status. (Judge Only)
func (h *TournamentHandler) DeleteTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can delete the tournament"})
		return
	}

	if t.Status == "playing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a tournament that is currently in progress. Finalize or rollback first."})
		return
	}

	if err := h.repo.DeleteTournament(c.Request.Context(), tournamentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament deleted successfully"})
}

// CompleteTournament marks a tournament as completed. (Judge Only)
func (h *TournamentHandler) CompleteTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uidStr, ok := uid.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal authentication error"})
		return
	}

	t, err := h.repo.GetTournament(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if t.CreatedBy != uidStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the tournament judge can finalize the tournament"})
		return
	}

	if t.Status != "playing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only finalize a tournament that is in progress"})
		return
	}

	t.Status = "completed"
	if err := h.repo.UpdateTournament(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament finalized successfully", "tournament": t})
}
