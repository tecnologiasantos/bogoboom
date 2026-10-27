package domain

import (
	"context"
	"time"
)

// Ports implemented by the PostgreSQL adapters (internal/adapters/repository). Services depend on
// these interfaces only, so tests use in-memory fakes.

// UserRepository stores accounts.
type UserRepository interface {
	GetByAuth0Sub(ctx context.Context, sub string) (*User, error) // nil, nil when absent
	GetByID(ctx context.Context, id string) (*User, error)        // nil, nil when absent
	// CreateWithProfile creates the user and the role's rows in one transaction.
	CreateNewcomer(ctx context.Context, u *User, n *Newcomer, c *NewcomerContact) error
	CreateHR(ctx context.Context, u *User, co *Company, fullName, phone string) error
	CreateAgency(ctx context.Context, u *User, a *Agency, fullName string) error
}

// NewcomerRepository stores newcomers and their private contacts.
type NewcomerRepository interface {
	// CreateInvited stores a newcomer created by HR (no account yet) with its private contact.
	CreateInvited(ctx context.Context, n *Newcomer, c *NewcomerContact) error
	GetByID(ctx context.Context, id string) (*Newcomer, error) // nil, nil when absent
	GetByUserID(ctx context.Context, userID string) (*Newcomer, error)
	ListByCompany(ctx context.Context, companyID string) ([]Newcomer, error)
	// ListOpen returns newcomers that allow contact, are not cancelled/moved in, newest first.
	ListOpen(ctx context.Context, limit int) ([]Newcomer, error)
	Update(ctx context.Context, n *Newcomer) error
	GetContact(ctx context.Context, newcomerID string) (*NewcomerContact, error)
}

// CompanyRepository reads the HR side.
type CompanyRepository interface {
	GetByHRUser(ctx context.Context, userID string) (*Company, error)
}

// AgencyRepository stores agencies, credits and unlocks.
type AgencyRepository interface {
	GetByUser(ctx context.Context, userID string) (*Agency, error)
	GetByID(ctx context.Context, id string) (*Agency, error)
	SetVerified(ctx context.Context, agencyID string, verified bool) error
	SetPlan(ctx context.Context, agencyID string, plan Plan, activeUntil *time.Time) error
	// ResetCreditsIfNewMonth refills the free credits when the stored period is older than period.
	ResetCreditsIfNewMonth(ctx context.Context, agencyID, period string, credits int) error
	// Unlock records the reveal and, when useCredit, decrements the credit atomically. It returns
	// ErrNoCredits when useCredit is true and none is left, and is a no-op for an existing unlock.
	Unlock(ctx context.Context, agencyID, newcomerID, userID string, useCredit bool) (alreadyUnlocked bool, err error)
	IsUnlocked(ctx context.Context, agencyID, newcomerID string) (bool, error)
	// UnlockedIDs returns the newcomer IDs the agency already unlocked.
	UnlockedIDs(ctx context.Context, agencyID string) (map[string]bool, error)
}

// PropertyRepository stores listings.
type PropertyRepository interface {
	Create(ctx context.Context, p *Property) error
	Update(ctx context.Context, p *Property) error // only if p.AgencyID owns it; ErrNotFound otherwise
	Delete(ctx context.Context, agencyID, id string) error
	GetByID(ctx context.Context, id string) (*Property, error)
	ListByAgency(ctx context.Context, agencyID string) ([]Property, error)
	// Search returns active listings of a city, optionally filtered by max price.
	Search(ctx context.Context, city string, maxPrice *float64, limit int) ([]Property, error)
	// ActiveByCity returns every active listing of a city (matching input).
	ActiveByCity(ctx context.Context, city string) ([]Property, error)
}

// FavoriteRepository stores a newcomer's saved listings.
type FavoriteRepository interface {
	Add(ctx context.Context, newcomerID, propertyID string) error
	Remove(ctx context.Context, newcomerID, propertyID string) error
	List(ctx context.Context, newcomerID string) ([]Property, error)
}

// NotificationRepository stores agency notifications.
type NotificationRepository interface {
	// Upsert creates the notification once per (agency, newcomer); it returns true when created.
	Upsert(ctx context.Context, n *Notification) (bool, error)
	ListByAgency(ctx context.Context, agencyID string, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, agencyID, id string) error
}

// InquiryRepository stores public form submissions.
type InquiryRepository interface {
	CreatePartnerInquiry(ctx context.Context, q *PartnerInquiry) error
	ListPartnerInquiries(ctx context.Context, limit int) ([]PartnerInquiry, error)
	CreateLandingLead(ctx context.Context, l *LandingLead) error
}

// AuditRepository stores the audit log.
type AuditRepository interface {
	Log(ctx context.Context, e AuditEntry) error
}
