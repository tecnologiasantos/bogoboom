package service

import (
	"context"
	"fmt"

	"bogoboom-backend/internal/core/domain"
)

// fanOutNewcomer notifies every agency whose listings match the newcomer well enough.
func fanOutNewcomer(ctx context.Context, props domain.PropertyRepository, notes domain.NotificationRepository, n domain.Newcomer) error {
	if !n.AllowContact || n.Stage == domain.StageCancelled || n.Stage == domain.StageMovedIn {
		return nil
	}
	list, err := props.ActiveByCity(ctx, n.CityOfContract)
	if err != nil {
		return err
	}
	byAgency := map[string][]domain.Property{}
	for _, p := range list {
		byAgency[p.AgencyID] = append(byAgency[p.AgencyID], p)
	}
	needs := domain.NeedsOf(n)
	for agencyID, ps := range byAgency {
		score := domain.BestScore(needs, ps)
		if score < domain.NotifyThreshold {
			continue
		}
		msg := fmt.Sprintf("%s is moving to %s and matches your listings (%d%%)", n.FirstName, n.CityOfContract, score)
		if _, err := notes.Upsert(ctx, &domain.Notification{AgencyID: agencyID, NewcomerID: n.ID, MatchScore: score, Message: msg}); err != nil {
			return err
		}
	}
	return nil
}

// fanOutProperty notifies one agency about the open newcomers that match a new listing.
func fanOutProperty(ctx context.Context, newcomers domain.NewcomerRepository, notes domain.NotificationRepository, p domain.Property) error {
	open, err := newcomers.ListOpen(ctx, 500)
	if err != nil {
		return err
	}
	facts := domain.FactsOf(p)
	for _, n := range open {
		if domain.Normalize(n.CityOfContract) != domain.Normalize(p.City) {
			continue
		}
		score := domain.ScoreMatch(domain.NeedsOf(n), facts)
		if score < domain.NotifyThreshold {
			continue
		}
		msg := fmt.Sprintf("%s is moving to %s and matches \"%s\" (%d%%)", n.FirstName, n.CityOfContract, p.Title, score)
		if _, err := notes.Upsert(ctx, &domain.Notification{AgencyID: p.AgencyID, NewcomerID: n.ID, MatchScore: score, Message: msg}); err != nil {
			return err
		}
	}
	return nil
}
