// Package domain holds the Bogoboom business entities, rules and ports. It imports the standard
// library only, so the rules can be tested without a database or a web framework.
//
// IDs are UUIDs carried as strings; dates without time are "YYYY-MM-DD" strings; money is float64
// (display and matching only, the database column is NUMERIC).
package domain

import "time"

// UserRole is what an account is allowed to do.
type UserRole string

const (
	RoleNewcomer UserRole = "newcomer"
	RoleHR       UserRole = "hr"
	RoleAgency   UserRole = "agency"
	RoleAdmin    UserRole = "admin"
)

// Valid reports whether r is one of the known roles.
func (r UserRole) Valid() bool {
	switch r {
	case RoleNewcomer, RoleHR, RoleAgency, RoleAdmin:
		return true
	}
	return false
}

// SelfRegistrable reports whether an account may pick this role by itself. Admin never can.
func (r UserRole) SelfRegistrable() bool {
	return r == RoleNewcomer || r == RoleHR || r == RoleAgency
}

// RelocationStage follows a newcomer from the HR invitation to moving in.
type RelocationStage string

const (
	StageInvited        RelocationStage = "invited"
	StageProfileDone    RelocationStage = "profile_done"
	StageSearching      RelocationStage = "searching"
	StageVisiting       RelocationStage = "visiting"
	StageContractSigned RelocationStage = "contract_signed"
	StageMovedIn        RelocationStage = "moved_in"
	StageCancelled      RelocationStage = "cancelled"
)

// Valid reports whether s is a known stage.
func (s RelocationStage) Valid() bool {
	switch s {
	case StageInvited, StageProfileDone, StageSearching, StageVisiting, StageContractSigned, StageMovedIn, StageCancelled:
		return true
	}
	return false
}

// Plan is an agency subscription level.
type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

// Valid reports whether p is a known plan.
func (p Plan) Valid() bool { return p == PlanFree || p == PlanPro || p == PlanEnterprise }

// PropertyType is the kind of dwelling.
type PropertyType string

const (
	TypeApartment PropertyType = "apartment"
	TypeRoom      PropertyType = "room"
	TypeHouse     PropertyType = "house"
	TypeStudio    PropertyType = "studio"
)

// Valid reports whether t is a known property type.
func (t PropertyType) Valid() bool {
	switch t {
	case TypeApartment, TypeRoom, TypeHouse, TypeStudio:
		return true
	}
	return false
}

// User is an authenticated account. Passwords never exist here: Auth0 owns authentication.
type User struct {
	ID        string    `json:"id"`
	Auth0Sub  string    `json:"-"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	Locale    string    `json:"locale"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

// Company is the employer an HR user works for.
type Company struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Industry           string `json:"industry"`
	City               string `json:"city"`
	Country            string `json:"country"`
	AllowAgencyContact bool   `json:"allowAgencyContact"`
}

// Agency is a real estate agency.
type Agency struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	City            string     `json:"city"`
	Country         string     `json:"country"`
	Description     string     `json:"description"`
	Verified        bool       `json:"verified"`
	ContactCredits  int        `json:"contactCredits"`
	CreditsPeriod   string     `json:"creditsPeriod"` // first day of the month the credits belong to, YYYY-MM-01
	Plan            Plan       `json:"plan"`
	PlanActiveUntil *time.Time `json:"planActiveUntil,omitempty"`
}

// Newcomer is the public ("blurred") part of a person who is relocating. Contact data lives in
// NewcomerContact and only leaves the server through an audited unlock.
type Newcomer struct {
	ID                 string          `json:"id"`
	UserID             *string         `json:"userId,omitempty"`
	CompanyID          *string         `json:"companyId,omitempty"`
	CreatedBy          *string         `json:"-"`
	FirstName          string          `json:"firstName"`
	CityOfContract     string          `json:"cityOfContract"`
	ArrivalDate        *string         `json:"arrivalDate,omitempty"`
	BudgetMin          *float64        `json:"budgetMin,omitempty"`
	BudgetMax          *float64        `json:"budgetMax,omitempty"`
	Currency           string          `json:"currency"`
	Bedrooms           *int            `json:"bedrooms,omitempty"`
	Furnished          *bool           `json:"furnished,omitempty"`
	PreferredDistricts []string        `json:"preferredDistricts"`
	Stage              RelocationStage `json:"stage"`
	AllowContact       bool            `json:"allowContact"`
	CreatedAt          time.Time       `json:"createdAt"`
}

// NewcomerContact is the private part of a newcomer.
type NewcomerContact struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// Property is a listing owned by an agency.
type Property struct {
	ID            string       `json:"id"`
	AgencyID      string       `json:"agencyId"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Price         float64      `json:"price"`
	Currency      string       `json:"currency"`
	Type          PropertyType `json:"type"`
	Bedrooms      *int         `json:"bedrooms,omitempty"`
	Furnished     bool         `json:"furnished"`
	City          string       `json:"city"`
	District      string       `json:"district"`
	Address       string       `json:"address"`
	AvailableFrom *string      `json:"availableFrom,omitempty"`
	Verified      bool         `json:"verified"`
	IsActive      bool         `json:"isActive"`
	CreatedAt     time.Time    `json:"createdAt"`
}

// Lead is what an agency sees about a newcomer. Contact is nil until the lead is unlocked.
type Lead struct {
	Newcomer
	MatchScore int              `json:"matchScore"`
	Unlocked   bool             `json:"unlocked"`
	Contact    *NewcomerContact `json:"contact,omitempty"`
}

// Notification tells an agency that a newcomer matches one of its listings.
type Notification struct {
	ID         string     `json:"id"`
	AgencyID   string     `json:"agencyId"`
	NewcomerID string     `json:"newcomerId"`
	MatchScore int        `json:"matchScore"`
	Message    string     `json:"message"`
	ReadAt     *time.Time `json:"readAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// PartnerInquiry comes from the public partnership form.
type PartnerInquiry struct {
	ID                string    `json:"id"`
	FullName          string    `json:"fullName"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Country           string    `json:"country"`
	City              string    `json:"city"`
	CollaborationType string    `json:"collaborationType"`
	CreatedAt         time.Time `json:"createdAt"`
}

// LandingLead comes from the public "keep me posted" form.
type LandingLead struct {
	Name     string
	Email    string
	Phone    string
	City     string
	Country  string
	Interest string
}

// AuditEntry records a sensitive action (contact reveal, plan change, verification, ...).
type AuditEntry struct {
	UserID    string
	Action    string
	TableName string
	RecordID  string
	Changes   map[string]any
	IP        string
}
