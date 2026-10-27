package service

import (
	"context"

	"bogoboom-backend/internal/core/domain"
)

// NewcomerUpdate is the body of PUT /newcomers/me. Only the fields that are sent change.
type NewcomerUpdate struct {
	CityOfContract *string   `json:"cityOfContract"`
	ArrivalDate    *string   `json:"arrivalDate"`
	BudgetMin      *float64  `json:"budgetMin"`
	BudgetMax      *float64  `json:"budgetMax"`
	Bedrooms       *int      `json:"bedrooms"`
	Furnished      *bool     `json:"furnished"`
	Districts      *[]string `json:"preferredDistricts"`
	Stage          *string   `json:"stage"`
	AllowContact   *bool     `json:"allowContact"`
}

// Invite is the body of POST /hr/newcomers: HR registers a new hire who is about to relocate.
type Invite struct {
	FullName       string   `json:"fullName"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone"`
	CityOfContract string   `json:"cityOfContract"`
	ArrivalDate    *string  `json:"arrivalDate"`
	BudgetMin      *float64 `json:"budgetMin"`
	BudgetMax      *float64 `json:"budgetMax"`
	Currency       string   `json:"currency"`
	Bedrooms       *int     `json:"bedrooms"`
	Furnished      *bool    `json:"furnished"`
	Districts      []string `json:"preferredDistricts"`
}

// NewcomerService covers what newcomers and HR do with newcomer profiles.
type NewcomerService struct {
	newcomers domain.NewcomerRepository
	companies domain.CompanyRepository
	props     domain.PropertyRepository
	notes     domain.NotificationRepository
	audit     domain.AuditRepository
}

// NewNewcomerService builds the service.
func NewNewcomerService(newcomers domain.NewcomerRepository, companies domain.CompanyRepository,
	props domain.PropertyRepository, notes domain.NotificationRepository, audit domain.AuditRepository) *NewcomerService {
	return &NewcomerService{newcomers: newcomers, companies: companies, props: props, notes: notes, audit: audit}
}

// Me returns the caller's own newcomer profile.
func (s *NewcomerService) Me(ctx context.Context, userID string) (*domain.Newcomer, error) {
	n, err := s.newcomers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, domain.ErrNotFound
	}
	return n, nil
}

// Update changes the caller's own profile. Becoming visible to agencies (allowContact) notifies the
// agencies whose listings match.
func (s *NewcomerService) Update(ctx context.Context, userID, ip string, u NewcomerUpdate) (*domain.Newcomer, error) {
	n, err := s.Me(ctx, userID)
	if err != nil {
		return nil, err
	}
	v := domain.NewValidationError()
	if u.CityOfContract != nil {
		c := domain.CleanText(*u.CityOfContract, 100)
		if c == "" {
			v.Add("cityOfContract", "cannot be blank")
		}
		n.CityOfContract = c
	}
	if u.ArrivalDate != nil {
		if *u.ArrivalDate == "" {
			n.ArrivalDate = nil
		} else if !domain.ValidDay(*u.ArrivalDate) {
			v.Add("arrivalDate", "must be YYYY-MM-DD")
		} else {
			n.ArrivalDate = u.ArrivalDate
		}
	}
	if u.BudgetMin != nil {
		n.BudgetMin = u.BudgetMin
	}
	if u.BudgetMax != nil {
		n.BudgetMax = u.BudgetMax
	}
	checkBudget(v, n.BudgetMin, n.BudgetMax)
	if u.Bedrooms != nil {
		if *u.Bedrooms < 0 || *u.Bedrooms > 20 {
			v.Add("bedrooms", "must be between 0 and 20")
		}
		n.Bedrooms = u.Bedrooms
	}
	if u.Furnished != nil {
		n.Furnished = u.Furnished
	}
	if u.Districts != nil {
		n.PreferredDistricts = domain.CleanList(*u.Districts, 10, 60)
	}
	if u.Stage != nil {
		st := domain.RelocationStage(*u.Stage)
		if !st.Valid() || st == domain.StageInvited {
			v.Add("stage", "is not a valid stage")
		}
		n.Stage = st
	}
	if u.AllowContact != nil {
		n.AllowContact = *u.AllowContact
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	if err := s.newcomers.Update(ctx, n); err != nil {
		return nil, err
	}
	if u.AllowContact != nil {
		_ = s.audit.Log(ctx, domain.AuditEntry{UserID: userID, Action: "newcomer.allow_contact", TableName: "newcomers", RecordID: n.ID,
			Changes: map[string]any{"allowContact": n.AllowContact}, IP: ip})
	}
	if n.AllowContact {
		if err := fanOutNewcomer(ctx, s.props, s.notes, *n); err != nil {
			return nil, err
		}
	}
	return n, nil
}

// Invite lets HR register a new hire. The newcomer is NOT visible to agencies until they sign up
// with the same e-mail and opt in: contact sharing is never decided by the employer.
func (s *NewcomerService) Invite(ctx context.Context, hrUserID, ip string, in Invite) (*domain.Newcomer, error) {
	company, err := s.companies.GetByHRUser(ctx, hrUserID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, domain.ErrForbidden
	}
	v := domain.NewValidationError()
	name := domain.CleanText(in.FullName, 120)
	if name == "" {
		v.Add("fullName", "is required")
	}
	email := domain.CleanText(in.Email, 254)
	if !domain.ValidEmail(email) {
		v.Add("email", "is not a valid e-mail")
	}
	phone := domain.CleanText(in.Phone, 20)
	if phone != "" && !domain.ValidPhone(phone) {
		v.Add("phone", "is not a valid phone number")
	}
	city := domain.CleanText(in.CityOfContract, 100)
	if city == "" {
		v.Add("cityOfContract", "is required")
	}
	if in.ArrivalDate != nil && !domain.ValidDay(*in.ArrivalDate) {
		v.Add("arrivalDate", "must be YYYY-MM-DD")
	}
	currency := in.Currency
	if currency == "" {
		currency = "COP"
	}
	if !domain.ValidCurrency(currency) {
		v.Add("currency", "must be a 3-letter code such as COP")
	}
	checkBudget(v, in.BudgetMin, in.BudgetMax)
	if in.Bedrooms != nil && (*in.Bedrooms < 0 || *in.Bedrooms > 20) {
		v.Add("bedrooms", "must be between 0 and 20")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	first := name
	for i, ch := range name {
		if ch == ' ' {
			first = name[:i]
			break
		}
	}
	n := &domain.Newcomer{
		CompanyID: &company.ID, CreatedBy: &hrUserID, FirstName: first, CityOfContract: city, ArrivalDate: in.ArrivalDate,
		BudgetMin: in.BudgetMin, BudgetMax: in.BudgetMax, Currency: currency, Bedrooms: in.Bedrooms, Furnished: in.Furnished,
		PreferredDistricts: domain.CleanList(in.Districts, 10, 60), Stage: domain.StageInvited, AllowContact: false,
	}
	if err := s.newcomers.CreateInvited(ctx, n, &domain.NewcomerContact{FullName: name, Email: email, Phone: phone}); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, domain.AuditEntry{UserID: hrUserID, Action: "hr.invite_newcomer", TableName: "newcomers", RecordID: n.ID, IP: ip})
	return n, nil
}

// ListForHR returns the newcomers of the caller's company.
func (s *NewcomerService) ListForHR(ctx context.Context, hrUserID string) ([]domain.Newcomer, error) {
	company, err := s.companies.GetByHRUser(ctx, hrUserID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, domain.ErrForbidden
	}
	return s.newcomers.ListByCompany(ctx, company.ID)
}
