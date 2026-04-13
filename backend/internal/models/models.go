package models

import "time"

// Player represents a participant in the tournament.
type Player struct {
	ID         string  `json:"id" firestore:"id"`
	Name       string  `json:"name" firestore:"name"`
	Email      string  `json:"email" firestore:"email"`
	TotalScore int     `json:"totalScore" firestore:"totalScore"`
	Wins       int     `json:"wins" firestore:"wins"`
	Losses     int     `json:"losses" firestore:"losses"`
	Draws      int     `json:"draws" firestore:"draws"`
	OMW        float64 `json:"omw" firestore:"omw"` // Opponent Match Win percentage
	GW         float64 `json:"gw" firestore:"gw"`   // Game Win percentage
	OGW        float64 `json:"ogw" firestore:"ogw"` // Opponent Game Win percentage
	Status     string  `json:"status" firestore:"status"` // "active", "dropped"
	HadBye     bool    `json:"hadBye" firestore:"hadBye"`
}

// Tournament represents the overall event.
type Tournament struct {
	ID           string    `json:"id" firestore:"id"`
	Name         string    `json:"name" firestore:"name"`
	Date         time.Time `json:"date" firestore:"date"`
	MaxPlayers   int       `json:"maxPlayers" firestore:"maxPlayers"`
	CurrentRound int       `json:"currentRound" firestore:"currentRound"`
	TotalRounds  int       `json:"totalRounds" firestore:"totalRounds"`
	CreatedBy    string    `json:"createdBy" firestore:"createdBy"`
	Status       string    `json:"status" firestore:"status"` // "registration", "playing", "completed"
	Format       string    `json:"format" firestore:"format"` // "BO1", "BO3"
	InviteCode   string    `json:"inviteCode" firestore:"inviteCode"`
}

// Round represents a specific round of the tournament.
type Round struct {
	ID           string    `json:"id" firestore:"id"`
	TournamentID string    `json:"tournamentId" firestore:"tournamentId"`
	RoundNumber  int       `json:"roundNumber" firestore:"roundNumber"`
	Status       string    `json:"status" firestore:"status"` // "pairing", "playing", "completed"
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
}

// Match represents a game between two players within a round.
type Match struct {
	ID           string `json:"id" firestore:"id"`
	RoundID      string `json:"roundId" firestore:"roundId"`
	Player1ID    string `json:"player1Id" firestore:"player1Id"`
	Player2ID    string `json:"player2Id" firestore:"player2Id"`
	Player1Score int    `json:"player1Score" firestore:"player1Score"`
	Player2Score int    `json:"player2Score" firestore:"player2Score"`
	WinnerID     string `json:"winnerId" firestore:"winnerId"`
	Status       string `json:"status" firestore:"status"` // "scheduled", "completed"
}

// Friendship represents a social link between two users.
type Friendship struct {
	ID        string    `json:"id" firestore:"id"`
	User1ID   string    `json:"user1Id" firestore:"user1Id"`
	User2ID   string    `json:"user2Id" firestore:"user2Id"`
	Status    string    `json:"status" firestore:"status"` // "pending", "accepted"
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}
