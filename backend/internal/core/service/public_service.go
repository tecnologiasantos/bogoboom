package service

import (
	"context"
	"time"

	"bogoboom-backend/internal/core/domain"
)

// InquiryInput is the body of POST /partner-inquiries (public).
type InquiryInput struct {
	FullName          string `json:"fullName"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Country           string `json:"country"`
	City              string `json:"city"`
	CollaborationType string `json:"collaborationType"`
	Consent           bool   `json:"acceptPrivacy"`
}

// LeadInput is the body of POST /leads (public).
type LeadInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Interest string `json:"interest"`
	Consent  bool   `json:"acceptPrivacy"`
}

// CollaborationTypes are the accepted values of the partnership form.
var CollaborationTypes = map[string]bool{"real_estate": true, "hr": true, "relocation": true, "franchise": true, "other": true}

// PublicService handles the unauthenticated forms and the admin tools around them.
type PublicService struct {
	inq    domain.InquiryRepository
	agency domain.AgencyRepository
	audit  domain.AuditRepository
}

// NewPublicService builds the service.
func NewPublicService(inq domain.InquiryRepository, agency domain.AgencyRepository, audit domain.AuditRepository) *PublicService {
	return &PublicService{inq: inq, agency: agency, audit: audit}
}

// CreatePartnerInquiry validates and stores a partnership request.
func (s *PublicService) CreatePartnerInquiry(ctx context.Context, in InquiryInput) error {
	v := domain.NewValidationError()
	q := &domain.PartnerInquiry{
		FullName: domain.CleanText(in.FullName, 120), Email: domain.CleanText(in.Email, 254), Phone: domain.CleanText(in.Phone, 20),
		Country: domain.CleanText(in.Country, 100), City: domain.CleanText(in.City, 100), CollaborationType: in.CollaborationType,
	}
	if q.FullName == "" {
		v.Add("fullName", "is required")
	}
	if !domain.ValidEmail(q.Email) {
		v.Add("email", "is not a valid e-mail")
	}
	if q.Phone != "" && !domain.ValidPhone(q.Phone) {
		v.Add("phone", "is not a valid phone number")
	}
	if !CollaborationTypes[q.CollaborationType] {
		v.Add("collaborationType", "is not a valid option")
	}
	if !in.Consent {
		v.Add("acceptPrivacy", "you must accept the privacy policy")
	}
	if err := v.Err(); err != nil {
		return err
	}
	return s.inq.CreatePartnerInquiry(ctx, q)
}

// CreateLead validates and stores a landing-page lead.
func (s *PublicService) CreateLead(ctx context.Context, in LeadInput) error {
	v := domain.NewValidationError()
	l := &domain.LandingLead{
		Name: domain.CleanText(in.Name, 120), Email: domain.CleanText(in.Email, 254), Phone: domain.CleanText(in.Phone, 20),
		City: domain.CleanText(in.City, 100), Country: domain.CleanText(in.Country, 100), Interest: domain.CleanText(in.Interest, 100),
	}
	if l.Name == "" {
		v.Add("name", "is required")
	}
	if !domain.ValidEmail(l.Email) {
		v.Add("email", "is not a valid e-mail")
	}
	if l.Phone != "" && !domain.ValidPhone(l.Phone) {
		v.Add("phone", "is not a valid phone number")
	}
	if !in.Consent {
		v.Add("acceptPrivacy", "you must accept the privacy policy")
	}
	if err := v.Err(); err != nil {
		return err
	}
	return s.inq.CreateLandingLead(ctx, l)
}

// ListInquiries is for admins.
func (s *PublicService) ListInquiries(ctx context.Context) ([]domain.PartnerInquiry, error) {
	return s.inq.ListPartnerInquiries(ctx, 200)
}

// VerifyAgency sets or clears the "verified" badge. Only admins call it.
func (s *PublicService) VerifyAgency(ctx context.Context, adminID, agencyID string, verified bool, ip string) error {
	a, err := s.agency.GetByID(ctx, agencyID)
	if err != nil {
		return err
	}
	if a == nil {
		return domain.ErrNotFound
	}
	if err := s.agency.SetVerified(ctx, agencyID, verified); err != nil {
		return err
	}
	return s.audit.Log(ctx, domain.AuditEntry{UserID: adminID, Action: "admin.verify_agency", TableName: "agencies", RecordID: agencyID,
		Changes: map[string]any{"verified": verified}, IP: ip})
}

// SetAgencyPlan changes an agency plan (until payments are automated, admins do it).
func (s *PublicService) SetAgencyPlan(ctx context.Context, adminID, agencyID string, plan domain.Plan, until *time.Time, ip string) error {
	if !plan.Valid() {
		v := domain.NewValidationError()
		v.Add("plan", "must be free, pro or enterprise")
		return v
	}
	a, err := s.agency.GetByID(ctx, agencyID)
	if err != nil {
		return err
	}
	if a == nil {
		return domain.ErrNotFound
	}
	if err := s.agency.SetPlan(ctx, agencyID, plan, until); err != nil {
		return err
	}
	return s.audit.Log(ctx, domain.AuditEntry{UserID: adminID, Action: "admin.set_plan", TableName: "agencies", RecordID: agencyID,
		Changes: map[string]any{"plan": string(plan)}, IP: ip})
}
