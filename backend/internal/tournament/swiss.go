package tournament

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/FranMaggi73/tcg-tournament/internal/models"
)

type SwissService struct {
	repo *Repository
}

func NewSwissService(repo *Repository) *SwissService {
	return &SwissService{repo: repo}
}

func (s *SwissService) CalculateTotalRounds(numPlayers int) int {
	if numPlayers <= 1 {
		return 0
	}
	return int(math.Ceil(math.Log2(float64(numPlayers))))
}

func (s *SwissService) GeneratePairings(ctx context.Context, tournamentID string) ([]*models.Match, error) {
	t, err := s.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	players, err := s.repo.GetPlayersByTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch players: %w", err)
	}

	var activePlayers []*models.Player
	for _, p := range players {
		if p.Status == "active" {
			activePlayers = append(activePlayers, p)
		}
	}

	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[i].TotalScore > activePlayers[j].TotalScore
	})

	// If it's the first round, shuffle the players since everyone has the same score
	if t.CurrentRound == 0 {
		rand.Shuffle(len(activePlayers), func(i, j int) {
			activePlayers[i], activePlayers[j] = activePlayers[j], activePlayers[i]
		})
	}

	previousMatches, err := s.getTournamentMatches(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match history: %w", err)
	}

	roundNum := t.CurrentRound + 1
	round := &models.Round{
		ID:           uuid.NewString(),
		TournamentID: tournamentID,
		RoundNumber:  roundNum,
		Status:       "playing",
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateRound(ctx, tournamentID, round); err != nil {
		return nil, fmt.Errorf("failed to create round document: %w", err)
	}

	var matches []*models.Match
	paired := make(map[string]bool)

	for i := 0; i < len(activePlayers); i++ {
		p1 := activePlayers[i]
		if paired[p1.ID] {
			continue
		}

		foundPair := false
		for j := i + 1; j < len(activePlayers); j++ {
			p2 := activePlayers[j]
			if paired[p2.ID] {
				continue
			}

			if s.havePlayedBefore(p1.ID, p2.ID, previousMatches) {
				continue
			}

			match := &models.Match{
				ID:        uuid.NewString(),
				RoundID:   round.ID,
				Player1ID: p1.ID,
				Player2ID: p2.ID,
				Status:    "scheduled",
			}
			matches = append(matches, match)
			paired[p1.ID] = true
			paired[p2.ID] = true
			foundPair = true
			break
		}

		if !foundPair {
			var byePlayer *models.Player
			for k := len(activePlayers) - 1; k >= 0; k-- {
				candidate := activePlayers[k]
				if !paired[candidate.ID] && !candidate.HadBye {
					byePlayer = candidate
					break
				}
			}

			if byePlayer == nil {
				byePlayer = p1
			}

			match := &models.Match{
				ID:        uuid.NewString(),
				RoundID:   round.ID,
				Player1ID: byePlayer.ID,
				Player2ID: "BYE",
				WinnerID:  byePlayer.ID,
				Status:    "completed",
			}
			matches = append(matches, match)
			paired[byePlayer.ID] = true
			byePlayer.HadBye = true
			s.repo.UpdatePlayer(ctx, tournamentID, byePlayer)
		}
	}

	for _, m := range matches {
		if err := s.repo.CreateMatch(ctx, tournamentID, round.ID, m); err != nil {
			return nil, fmt.Errorf("failed to create match document: %w", err)
		}
	}

	t.CurrentRound = roundNum
	t.Status = "playing"
	if err := s.repo.UpdateTournament(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to update tournament state: %w", err)
	}

	return matches, nil
}

func (s *SwissService) havePlayedBefore(p1, p2 string, matches []*models.Match) bool {
	for _, m := range matches {
		if (m.Player1ID == p1 && m.Player2ID == p2) || (m.Player1ID == p2 && m.Player2ID == p1) {
			return true
		}
	}
	return false
}

func (s *SwissService) getTournamentMatches(ctx context.Context, tournamentID string) ([]*models.Match, error) {
	return s.repo.GetAllMatches(ctx, tournamentID)
}

func (s *SwissService) ProcessMatchResult(ctx context.Context, tournamentID string, roundID string, matchID string, match *models.Match) error {
	p1Score, p1Wins, p1Loss, p1Draw := 0, 0, 0, 0
	p2Score, p2Wins, p2Loss, p2Draw := 0, 0, 0, 0

	if match.Player2ID == "BYE" {
		p1Score, p1Wins = 3, 1
	} else if match.WinnerID == match.Player1ID {
		p1Score, p1Wins = 3, 1
		p2Loss = 1
	} else if match.WinnerID == match.Player2ID {
		p2Score, p2Wins = 3, 1
		p1Loss = 1
	} else {
		p1Score, p2Score = 1, 1
		p1Draw, p2Draw = 1, 1
	}

	return s.repo.ProcessMatchAtomic(ctx, tournamentID, roundID, matchID, match, p1Score, p1Wins, p1Loss, p1Draw, p2Score, p2Wins, p2Loss, p2Draw)
}

func (s *SwissService) UpdateStandings(ctx context.Context, tournamentID string) error {
	players, err := s.repo.GetPlayersByTournament(ctx, tournamentID)
	if err != nil {
		return err
	}

	matches, err := s.getTournamentMatches(ctx, tournamentID)
	if err != nil {
		return err
	}

	for _, p := range players {
		var totalGames, gamesWon float64
		for _, m := range matches {
			if m.Player1ID == p.ID || m.Player2ID == p.ID {
				if m.Player2ID == "BYE" && m.Player1ID == p.ID {
					continue
				}
				totalGames += float64(m.Player1Score + m.Player2Score)
				if m.Player1ID == p.ID {
					gamesWon += float64(m.Player1Score)
				} else {
					gamesWon += float64(m.Player2Score)
				}
			}
		}
		if totalGames > 0 {
			p.GW = gamesWon / totalGames
		}

		var opponentWins, opponentMatches float64
		for _, m := range matches {
			var oppID string
			if m.Player1ID == p.ID {
				oppID = m.Player2ID
			} else if m.Player2ID == p.ID {
				oppID = m.Player1ID
			} else {
				continue
			}
			if oppID == "BYE" {
				continue
			}

			opp, err := s.repo.GetPlayer(ctx, tournamentID, oppID)
			if err == nil && opp != nil {
				opponentWins += float64(opp.Wins)
				opponentMatches += float64(opp.Wins + opp.Losses + opp.Draws)
			}
		}
		if opponentMatches > 0 {
			p.OMW = opponentWins / opponentMatches
		}

		var sumOGW float64
		var oppCount float64
		for _, m := range matches {
			var oppID string
			if m.Player1ID == p.ID {
				oppID = m.Player2ID
			} else if m.Player2ID == p.ID {
				oppID = m.Player1ID
			} else {
				continue
			}
			if oppID == "BYE" {
				continue
			}

			opp, err := s.repo.GetPlayer(ctx, tournamentID, oppID)
			if err == nil && opp != nil {
				sumOGW += opp.GW
				oppCount++
			}
		}
		if oppCount > 0 {
			p.OGW = sumOGW / oppCount
		}

		s.repo.UpdatePlayer(ctx, tournamentID, p)
	}
	return nil
}
