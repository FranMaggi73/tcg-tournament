package tournament

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"github.com/FranMaggi73/tcg-tournament/backend/internal/models"
)

type Repository struct {
	client *firestore.Client
}

func NewRepository(client *firestore.Client) *Repository {
	return &Repository{
		client: client,
	}
}

// --- Tournament Methods ---

func (r *Repository) CreateTournament(ctx context.Context, t *models.Tournament) error {
	_, err := r.client.Collection("tournaments").Doc(t.ID).Set(ctx, t)
	return err
}

func (r *Repository) GetTournament(ctx context.Context, id string) (*models.Tournament, error) {
	doc, err := r.client.Collection("tournaments").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var t models.Tournament
	doc.DataTo(&t)
	return &t, nil
}

func (r *Repository) UpdateTournament(ctx context.Context, t *models.Tournament) error {
	_, err := r.client.Collection("tournaments").Doc(t.ID).Set(ctx, t)
	return err
}

func (r *Repository) GetTournamentByInviteCode(ctx context.Context, code string) (*models.Tournament, error) {
	iter := r.client.Collection("tournaments").Where("inviteCode", "==", code).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("no tournament found with this invite code")
	}
	if err != nil {
		return nil, err
	}
	var t models.Tournament
	doc.DataTo(&t)
	return &t, nil
}

func (r *Repository) DeleteTournament(ctx context.Context, id string) error {
	_, err := r.client.Collection("tournaments").Doc(id).Delete(ctx)
	return err
}

// --- Player Methods ---

func (r *Repository) PlayerExists(ctx context.Context, tournamentID string, email string) (bool, error) {
	iter := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("players").Where("email", "==", email).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return doc != nil, nil
}

func (r *Repository) CreatePlayer(ctx context.Context, tournamentID string, p *models.Player) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("players").Doc(p.ID).Set(ctx, p)
	return err
}

func (r *Repository) GetPlayer(ctx context.Context, tournamentID string, playerID string) (*models.Player, error) {
	doc, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("players").Doc(playerID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var p models.Player
	doc.DataTo(&p)
	return &p, nil
}

func (r *Repository) GetPlayersByTournament(ctx context.Context, tournamentID string) ([]*models.Player, error) {
	iter := r.client.Collection("tournaments").Doc(tournamentID).Collection("players").Documents(ctx)
	var players []*models.Player
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var p models.Player
		doc.DataTo(&p)
		players = append(players, &p)
	}
	return players, nil
}

func (r *Repository) UpdatePlayer(ctx context.Context, tournamentID string, p *models.Player) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("players").Doc(p.ID).Set(ctx, p)
	return err
}

func (r *Repository) UpdatePlayerStatus(ctx context.Context, tournamentID string, playerID string, status string) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("players").Doc(playerID).Update(ctx, []firestore.Update{
			{Path: "status", Value: status},
		})
	return err
}

// --- Friendship Methods ---

func (r *Repository) CreateFriendship(ctx context.Context, f *models.Friendship) error {
	_, err := r.client.Collection("friendships").Doc(f.ID).Set(ctx, f)
	return err
}

func (r *Repository) GetFriends(ctx context.Context, userID string) ([]*models.Friendship, error) {
	var allFriends []*models.Friendship

	// Query for user1
	iter1 := r.client.Collection("friendships").Where("user1Id", "==", userID).Where("status", "==", "accepted").Documents(ctx)
	for {
		doc, err := iter1.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		var f models.Friendship
		doc.DataTo(&f)
		allFriends = append(allFriends, &f)
	}

	// Query for user2
	iter2 := r.client.Collection("friendships").Where("user2Id", "==", userID).Where("status", "==", "accepted").Documents(ctx)
	for {
		doc, err := iter2.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		var f models.Friendship
		doc.DataTo(&f)
		allFriends = append(allFriends, &f)
	}

	return allFriends, nil
}

func (r *Repository) UpdateFriendshipStatus(ctx context.Context, friendshipID string, status string) error {
	_, err := r.client.Collection("friendships").Doc(friendshipID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
	})
	return err
}

func (r *Repository) GetPendingRequests(ctx context.Context, userID string) ([]*models.Friendship, error) {
	iter := r.client.Collection("friendships").Where("user2Id", "==", userID).Where("status", "==", "pending").Documents(ctx)
	var requests []*models.Friendship
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var f models.Friendship
		doc.DataTo(&f)
		requests = append(requests, &f)
	}
	return requests, nil
}

// --- Round Methods ---

func (r *Repository) CreateRound(ctx context.Context, tournamentID string, rnd *models.Round) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(rnd.ID).Set(ctx, rnd)
	return err
}

func (r *Repository) GetRoundsByTournament(ctx context.Context, tournamentID string) ([]*models.Round, error) {
	iter := r.client.Collection("tournaments").Doc(tournamentID).Collection("rounds").Documents(ctx)
	var rounds []*models.Round
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var rnd models.Round
		doc.DataTo(&rnd)
		rounds = append(rounds, &rnd)
	}
	return rounds, nil
}

func (r *Repository) DeleteRound(ctx context.Context, tournamentID string, roundID string) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(roundID).Delete(ctx)
	return err
}

// --- Match Methods ---

func (r *Repository) GetMatch(ctx context.Context, tournamentID string, roundID string, matchID string) (*models.Match, error) {
	doc, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(roundID).
		Collection("matches").Doc(matchID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m models.Match
	doc.DataTo(&m)
	return &m, nil
}

func (r *Repository) CreateMatch(ctx context.Context, tournamentID string, roundID string, m *models.Match) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(roundID).
		Collection("matches").Doc(m.ID).Set(ctx, m)
	return err
}

func (r *Repository) GetMatchesByRound(ctx context.Context, tournamentID string, roundID string) ([]*models.Match, error) {
	iter := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(roundID).
		Collection("matches").Documents(ctx)
	var matches []*models.Match
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m models.Match
		doc.DataTo(&m)
		matches = append(matches, &m)
	}
	return matches, nil
}

func (r *Repository) GetAllMatches(ctx context.Context, tournamentID string) ([]*models.Match, error) {
	var allMatches []*models.Match
	roundsIter := r.client.Collection("tournaments").Doc(tournamentID).Collection("rounds").Documents(ctx)

	for {
		roundDoc, err := roundsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		matchIter := r.client.Collection("tournaments").Doc(tournamentID).
			Collection("rounds").Doc(roundDoc.Ref.ID).
			Collection("matches").Documents(ctx)

		for {
			matchDoc, err := matchIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}
			var m models.Match
			matchDoc.DataTo(&m)
			allMatches = append(allMatches, &m)
		}
	}
	return allMatches, nil
}

func (r *Repository) ProcessMatchAtomic(ctx context.Context, tournamentID string, roundID string, matchID string, match *models.Match, p1Score, p1Wins, p1Loss, p1Draw, p2Score, p2Wins, p2Loss, p2Draw int) error {
	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		matchRef := r.client.Collection("tournaments").Doc(tournamentID).
			Collection("rounds").Doc(roundID).
			Collection("matches").Doc(matchID)

		if err := tx.Set(matchRef, match); err != nil {
			return err
		}

		if match.Player1ID != "" && match.Player1ID != "BYE" {
			p1Ref := r.client.Collection("tournaments").Doc(tournamentID).Collection("players").Doc(match.Player1ID)
			p1Doc, err := tx.Get(p1Ref)
			if err != nil {
				return err
			}
			var p1 models.Player
			p1Doc.DataTo(&p1)
			p1.TotalScore += p1Score
			p1.Wins += p1Wins
			p1.Losses += p1Loss
			p1.Draws += p1Draw
			if err := tx.Set(p1Ref, p1); err != nil {
				return err
			}
		}

		if match.Player2ID != "" && match.Player2ID != "BYE" {
			p2Ref := r.client.Collection("tournaments").Doc(tournamentID).Collection("players").Doc(match.Player2ID)
			p2Doc, err := tx.Get(p2Ref)
			if err != nil {
				return err
			}
			var p2 models.Player
			p2Doc.DataTo(&p2)
			p2.TotalScore += p2Score
			p2.Wins += p2Wins
			p2.Losses += p2Loss
			p2.Draws += p2Draw
			if err := tx.Set(p2Ref, p2); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) DeleteMatch(ctx context.Context, tournamentID string, roundID string, matchID string) error {
	_, err := r.client.Collection("tournaments").Doc(tournamentID).
		Collection("rounds").Doc(roundID).
		Collection("matches").Doc(matchID).Delete(ctx)
	return err
}
