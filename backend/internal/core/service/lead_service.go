package service

import (
	"context"
	"sort"
	"time"

	"bogoboom-backend/internal/core/domain"
)

// LeadService is the agency's view of newcomers. Contact data leaves the server only through
// Unlock, which checks verification, consent, match and credits, and writes an audit entry.
type LeadService struct {
	agencies  domain.AgencyRepository
	newcomers domain.NewcomerRepository
	props     domain.PropertyRepository
	notes     domain.NotificationRepository
	audit     domain.AuditRepository
	clock     Clock
}

// NewLeadService builds the service.
func NewLeadService(a domain.AgencyRepository, n domain.NewcomerRepository, p domain.PropertyRepository,
	notes domain.NotificationRepository, audit domain.AuditRepository) *LeadService {
	return &LeadService{agencies: a, newcomers: n, props: p, notes: notes, audit: audit}
}

// PlanActive reports whether the agency has a paid plan that has not expired.
func PlanActive(a *domain.Agency, now time.Time) bool {
	if a.Plan != domain.PlanPro && a.Plan != domain.PlanEnterprise {
		return false
	}
	return a.PlanActiveUntil == nil || a.PlanActiveUntil.After(now)
}

func (s *LeadService) agencyOf(ctx context.Context, userID string) (*domain.Agency, error) {
	a, err := s.agencies.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, domain.ErrForbidden
	}
	// Lazy monthly refill: nobody has to run a cron job.
	period := MonthStart(s.clock.now())
	if a.CreditsPeriod < period {
		if err := s.agencies.ResetCreditsIfNewMonth(ctx, a.ID, period, FreeMonthlyCredits); err != nil {
			return nil, err
		}
		a.CreditsPeriod, a.ContactCredits = period, FreeMonthlyCredits
	}
	return a, nil
}

// Account returns the agency of the caller with its credits refreshed.
func (s *LeadService) Account(ctx context.Context, userID string) (*domain.Agency, error) {
	return s.agencyOf(ctx, userID)
}

// List returns the newcomers that match the agency's listings, best match first. The result is
// "blurred": no contact data, except for leads the agency already unlocked.
func (s *LeadService) List(ctx context.Context, userID string) ([]domain.Lead, error) {
	a, err := s.agencyOf(ctx, userID)
	if err != nil {
		return nil, err
	}
	mine, err := s.props.ListByAgency(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	active := mine[:0:0]
	for _, p := range mine {
		if p.IsActive {
			active = append(active, p)
		}
	}
	open, err := s.newcomers.ListOpen(ctx, 500)
	if err != nil {
		return nil, err
	}
	unlocked, err := s.agencies.UnlockedIDs(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	leads := []domain.Lead{}
	for _, n := range open {
		score := domain.BestScore(domain.NeedsOf(n), active)
		if score < domain.NotifyThreshold {
			continue
		}
		l := domain.Lead{Newcomer: n, MatchScore: score, Unlocked: unlocked[n.ID]}
		if l.Unlocked {
			if c, err := s.newcomers.GetContact(ctx, n.ID); err == nil {
				l.Contact = c
			}
		}
		leads = append(leads, l)
	}
	sort.SliceStable(leads, func(i, j int) bool { return leads[i].MatchScore > leads[j].MatchScore })
	return leads, nil
}

// Unlock reveals the contact of a lead. Rules, in order: the agency is verified, the newcomer opted
// in, the newcomer matches one of the agency's active listings, and the agency has a paid plan or a
// free credit left. Unlocking the same lead twice is free and idempotent.
func (s *LeadService) Unlock(ctx context.Context, userID, newcomerID, ip string) (*domain.Lead, error) {
	a, err := s.agencyOf(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !a.Verified {
		return nil, domain.ErrAgencyNotVerified
	}
	n, err := s.newcomers.GetByID(ctx, newcomerID)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, domain.ErrNotFound
	}
	if !n.AllowContact || n.Stage == domain.StageCancelled {
		return nil, domain.ErrContactNotAllowed
	}
	mine, err := s.props.ListByAgency(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	active := mine[:0:0]
	for _, p := range mine {
		if p.IsActive {
			active = append(active, p)
		}
	}
	score := domain.BestScore(domain.NeedsOf(*n), active)
	if score < domain.NotifyThreshold {
		return nil, domain.ErrNotMatching
	}
	useCredit := !PlanActive(a, s.clock.now())
	if useCredit && a.ContactCredits <= 0 {
		already, err := s.agencies.IsUnlocked(ctx, a.ID, n.ID)
		if err != nil {
			return nil, err
		}
		if !already {
			return nil, domain.ErrNoCredits
		}
	}
	already, err := s.agencies.Unlock(ctx, a.ID, n.ID, userID, useCredit)
	if err != nil {
		return nil, err
	}
	contact, err := s.newcomers.GetContact(ctx, n.ID)
	if err != nil {
		return nil, err
	}
	if !already {
		_ = s.audit.Log(ctx, domain.AuditEntry{UserID: userID, Action: "agency.unlock_contact", TableName: "newcomers", RecordID: n.ID,
			Changes: map[string]any{"agencyId": a.ID, "usedCredit": useCredit, "score": score}, IP: ip})
	}
	return &domain.Lead{Newcomer: *n, MatchScore: score, Unlocked: true, Contact: contact}, nil
}

// Notifications lists the agency's notifications.
func (s *LeadService) Notifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	a, err := s.agencyOf(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.notes.ListByAgency(ctx, a.ID, 100)
}

// MarkRead marks one notification of the agency as read.
func (s *LeadService) MarkRead(ctx context.Context, userID, id string) error {
	a, err := s.agencyOf(ctx, userID)
	if err != nil {
		return err
	}
	return s.notes.MarkRead(ctx, a.ID, id)
}

// WithClock replaces the clock (tests).
func (s *LeadService) WithClock(c Clock) *LeadService { s.clock = c; return s }
