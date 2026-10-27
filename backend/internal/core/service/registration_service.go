package service

import (
	"context"
	"strings"

	"bogoboom-backend/internal/core/domain"
)

// Identity is who the access token says the caller is.
type Identity struct {
	Sub   string
	Email string
}

// Consent must be true: the person accepted the terms and the privacy policy.
type Consent struct {
	Accepted bool `json:"acceptTerms"`
	// ShareWithAgencies is the opt-in that lets agencies see a newcomer's blurred profile.
	ShareWithAgencies bool `json:"shareWithAgencies"`
}

// NewcomerRegistration is the body of POST /register/newcomer.
type NewcomerRegistration struct {
	Consent
	FullName       string   `json:"fullName"`
	Phone          string   `json:"phone"`
	CityOfContract string   `json:"cityOfContract"`
	ArrivalDate    *string  `json:"arrivalDate"`
	BudgetMin      *float64 `json:"budgetMin"`
	BudgetMax      *float64 `json:"budgetMax"`
	Currency       string   `json:"currency"`
	Bedrooms       *int     `json:"bedrooms"`
	Furnished      *bool    `json:"furnished"`
	Districts      []string `json:"preferredDistricts"`
	Locale         string   `json:"locale"`
}

// HRRegistration is the body of POST /register/hr.
type HRRegistration struct {
	Consent
	FullName    string `json:"fullName"`
	Phone       string `json:"phone"`
	CompanyName string `json:"companyName"`
	Industry    string `json:"industry"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Locale      string `json:"locale"`
}

// AgencyRegistration is the body of POST /register/agency.
type AgencyRegistration struct {
	Consent
	FullName    string `json:"fullName"`
	AgencyName  string `json:"agencyName"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Description string `json:"description"`
	Locale      string `json:"locale"`
}

// RegistrationService creates accounts for an authenticated identity. The role comes from the
// endpoint, never from a free field, and admin can never be self-registered.
type RegistrationService struct {
	users domain.UserRepository
	clock Clock
}

// NewRegistrationService builds the service.
func NewRegistrationService(users domain.UserRepository) *RegistrationService {
	return &RegistrationService{users: users}
}

func (s *RegistrationService) precheck(ctx context.Context, id Identity, c Consent) (string, error) {
	email := strings.ToLower(domain.CleanText(id.Email, 254))
	if id.Sub == "" || !domain.ValidEmail(email) {
		return "", domain.ErrEmailRequired
	}
	if !c.Accepted {
		return "", domain.ErrConsentRequired
	}
	existing, err := s.users.GetByAuth0Sub(ctx, id.Sub)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", domain.ErrAlreadyRegistered
	}
	return email, nil
}

func localeOr(l string) string {
	switch l {
	case "es", "pt":
		return l
	}
	return "en"
}

func checkBudget(v *domain.ValidationError, min, max *float64) {
	if min != nil && (*min < 0 || *min > 1e12) {
		v.Add("budgetMin", "must be between 0 and 1,000,000,000,000")
	}
	if max != nil && (*max < 0 || *max > 1e12) {
		v.Add("budgetMax", "must be between 0 and 1,000,000,000,000")
	}
	if min != nil && max != nil && *min > *max {
		v.Add("budgetMax", "must be greater than or equal to the minimum")
	}
}

// RegisterNewcomer creates a newcomer account. If HR already invited this e-mail, the repository
// links the account to that invitation instead of creating a second profile.
func (s *RegistrationService) RegisterNewcomer(ctx context.Context, id Identity, r NewcomerRegistration) (*domain.User, error) {
	email, err := s.precheck(ctx, id, r.Consent)
	if err != nil {
		return nil, err
	}
	v := domain.NewValidationError()
	name := domain.CleanText(r.FullName, 120)
	if name == "" {
		v.Add("fullName", "is required")
	}
	phone := domain.CleanText(r.Phone, 20)
	if phone != "" && !domain.ValidPhone(phone) {
		v.Add("phone", "is not a valid phone number")
	}
	city := domain.CleanText(r.CityOfContract, 100)
	if city == "" {
		v.Add("cityOfContract", "is required")
	}
	if r.ArrivalDate != nil && !domain.ValidDay(*r.ArrivalDate) {
		v.Add("arrivalDate", "must be YYYY-MM-DD")
	}
	currency := r.Currency
	if currency == "" {
		currency = "COP"
	}
	if !domain.ValidCurrency(currency) {
		v.Add("currency", "must be a 3-letter code such as COP")
	}
	checkBudget(v, r.BudgetMin, r.BudgetMax)
	if r.Bedrooms != nil && (*r.Bedrooms < 0 || *r.Bedrooms > 20) {
		v.Add("bedrooms", "must be between 0 and 20")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}

	user := &domain.User{Auth0Sub: id.Sub, Email: email, Role: domain.RoleNewcomer, Locale: localeOr(r.Locale), IsActive: true}
	first := name
	for i, ch := range name {
		if ch == ' ' {
			first = name[:i]
			break
		}
	}
	n := &domain.Newcomer{
		FirstName: first, CityOfContract: city, ArrivalDate: r.ArrivalDate,
		BudgetMin: r.BudgetMin, BudgetMax: r.BudgetMax, Currency: currency,
		Bedrooms: r.Bedrooms, Furnished: r.Furnished,
		PreferredDistricts: domain.CleanList(r.Districts, 10, 60),
		Stage:              domain.StageProfileDone, AllowContact: r.ShareWithAgencies,
	}
	if err := s.users.CreateNewcomer(ctx, user, n, &domain.NewcomerContact{FullName: name, Email: email, Phone: phone}); err != nil {
		return nil, err
	}
	return user, nil
}

// RegisterHR creates an HR account together with its company.
func (s *RegistrationService) RegisterHR(ctx context.Context, id Identity, r HRRegistration) (*domain.User, error) {
	email, err := s.precheck(ctx, id, r.Consent)
	if err != nil {
		return nil, err
	}
	v := domain.NewValidationError()
	name := domain.CleanText(r.FullName, 120)
	if name == "" {
		v.Add("fullName", "is required")
	}
	phone := domain.CleanText(r.Phone, 20)
	if phone != "" && !domain.ValidPhone(phone) {
		v.Add("phone", "is not a valid phone number")
	}
	company := domain.CleanText(r.CompanyName, 160)
	if company == "" {
		v.Add("companyName", "is required")
	}
	country := domain.CleanText(r.Country, 10)
	if len(country) != 2 {
		v.Add("country", "must be a 2-letter ISO code such as CO")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	user := &domain.User{Auth0Sub: id.Sub, Email: email, Role: domain.RoleHR, Locale: localeOr(r.Locale), IsActive: true}
	co := &domain.Company{Name: company, Industry: domain.CleanText(r.Industry, 100), City: domain.CleanText(r.City, 100),
		Country: upper(country), AllowAgencyContact: false}
	if err := s.users.CreateHR(ctx, user, co, name, phone); err != nil {
		return nil, err
	}
	return user, nil
}

// RegisterAgency creates an agency account together with the agency (unverified, free plan).
func (s *RegistrationService) RegisterAgency(ctx context.Context, id Identity, r AgencyRegistration) (*domain.User, error) {
	email, err := s.precheck(ctx, id, r.Consent)
	if err != nil {
		return nil, err
	}
	v := domain.NewValidationError()
	name := domain.CleanText(r.FullName, 120)
	if name == "" {
		v.Add("fullName", "is required")
	}
	agency := domain.CleanText(r.AgencyName, 160)
	if agency == "" {
		v.Add("agencyName", "is required")
	}
	city := domain.CleanText(r.City, 100)
	if city == "" {
		v.Add("city", "is required")
	}
	country := domain.CleanText(r.Country, 10)
	if len(country) != 2 {
		v.Add("country", "must be a 2-letter ISO code such as CO")
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	user := &domain.User{Auth0Sub: id.Sub, Email: email, Role: domain.RoleAgency, Locale: localeOr(r.Locale), IsActive: true}
	a := &domain.Agency{Name: agency, City: city, Country: upper(country), Description: domain.CleanMultiline(r.Description, 2000),
		Plan: domain.PlanFree, ContactCredits: FreeMonthlyCredits, CreditsPeriod: MonthStart(s.clock.now())}
	if err := s.users.CreateAgency(ctx, user, a, name); err != nil {
		return nil, err
	}
	return user, nil
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}

// WithClock replaces the clock (tests).
func (s *RegistrationService) WithClock(c Clock) *RegistrationService { s.clock = c; return s }
