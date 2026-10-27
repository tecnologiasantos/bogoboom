package service

import (
	"context"

	"bogoboom-backend/internal/core/domain"
)

// PropertyInput is the body of POST/PUT /agency/properties.
type PropertyInput struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	Type          string  `json:"type"`
	Bedrooms      *int    `json:"bedrooms"`
	Furnished     bool    `json:"furnished"`
	City          string  `json:"city"`
	District      string  `json:"district"`
	Address       string  `json:"address"`
	AvailableFrom *string `json:"availableFrom"`
	IsActive      *bool   `json:"isActive"`
}

// PropertyService manages listings. Agencies manage their own; newcomers search the public view.
type PropertyService struct {
	agencies  domain.AgencyRepository
	props     domain.PropertyRepository
	newcomers domain.NewcomerRepository
	notes     domain.NotificationRepository
	favs      domain.FavoriteRepository
}

// NewPropertyService builds the service.
func NewPropertyService(a domain.AgencyRepository, p domain.PropertyRepository, n domain.NewcomerRepository,
	notes domain.NotificationRepository, favs domain.FavoriteRepository) *PropertyService {
	return &PropertyService{agencies: a, props: p, newcomers: n, notes: notes, favs: favs}
}

func (s *PropertyService) agency(ctx context.Context, userID string) (*domain.Agency, error) {
	a, err := s.agencies.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, domain.ErrForbidden
	}
	return a, nil
}

func buildProperty(in PropertyInput, p *domain.Property) error {
	v := domain.NewValidationError()
	p.Title = domain.CleanText(in.Title, 160)
	if p.Title == "" {
		v.Add("title", "is required")
	}
	p.Description = domain.CleanMultiline(in.Description, 4000)
	if in.Price < 0 || in.Price > 1e12 {
		v.Add("price", "must be between 0 and 1,000,000,000,000")
	}
	p.Price = in.Price
	p.Currency = in.Currency
	if p.Currency == "" {
		p.Currency = "COP"
	}
	if !domain.ValidCurrency(p.Currency) {
		v.Add("currency", "must be a 3-letter code such as COP")
	}
	p.Type = domain.PropertyType(in.Type)
	if !p.Type.Valid() {
		v.Add("type", "must be apartment, room, house or studio")
	}
	if in.Bedrooms != nil && (*in.Bedrooms < 0 || *in.Bedrooms > 20) {
		v.Add("bedrooms", "must be between 0 and 20")
	}
	p.Bedrooms = in.Bedrooms
	p.Furnished = in.Furnished
	p.City = domain.CleanText(in.City, 100)
	if p.City == "" {
		v.Add("city", "is required")
	}
	p.District = domain.CleanText(in.District, 100)
	p.Address = domain.CleanText(in.Address, 200)
	if in.AvailableFrom != nil && *in.AvailableFrom != "" {
		if !domain.ValidDay(*in.AvailableFrom) {
			v.Add("availableFrom", "must be YYYY-MM-DD")
		}
		p.AvailableFrom = in.AvailableFrom
	} else {
		p.AvailableFrom = nil
	}
	if in.IsActive != nil {
		p.IsActive = *in.IsActive
	}
	return v.Err()
}

// Create adds a listing and notifies the agency about newcomers that match it.
func (s *PropertyService) Create(ctx context.Context, userID string, in PropertyInput) (*domain.Property, error) {
	a, err := s.agency(ctx, userID)
	if err != nil {
		return nil, err
	}
	p := &domain.Property{AgencyID: a.ID, IsActive: true}
	if err := buildProperty(in, p); err != nil {
		return nil, err
	}
	if err := s.props.Create(ctx, p); err != nil {
		return nil, err
	}
	if p.IsActive {
		if err := fanOutProperty(ctx, s.newcomers, s.notes, *p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Update replaces a listing of the caller's agency.
func (s *PropertyService) Update(ctx context.Context, userID, id string, in PropertyInput) (*domain.Property, error) {
	a, err := s.agency(ctx, userID)
	if err != nil {
		return nil, err
	}
	p, err := s.props.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil || p.AgencyID != a.ID {
		return nil, domain.ErrNotFound
	}
	if err := buildProperty(in, p); err != nil {
		return nil, err
	}
	if err := s.props.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete removes a listing of the caller's agency.
func (s *PropertyService) Delete(ctx context.Context, userID, id string) error {
	a, err := s.agency(ctx, userID)
	if err != nil {
		return err
	}
	return s.props.Delete(ctx, a.ID, id)
}

// ListMine returns the caller's agency listings.
func (s *PropertyService) ListMine(ctx context.Context, userID string) ([]domain.Property, error) {
	a, err := s.agency(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.props.ListByAgency(ctx, a.ID)
}

// Public is a listing as newcomers see it: the exact street address stays private.
func Public(p domain.Property) domain.Property {
	p.Address = ""
	return p
}

// Search returns active listings in a city for newcomers, without street addresses.
func (s *PropertyService) Search(ctx context.Context, city string, maxPrice *float64) ([]domain.Property, error) {
	city = domain.CleanText(city, 100)
	if city == "" {
		v := domain.NewValidationError()
		v.Add("city", "is required")
		return nil, v
	}
	list, err := s.props.Search(ctx, city, maxPrice, 100)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Property, len(list))
	for i, p := range list {
		out[i] = Public(p)
	}
	return out, nil
}

func (s *PropertyService) newcomer(ctx context.Context, userID string) (*domain.Newcomer, error) {
	n, err := s.newcomers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, domain.ErrForbidden
	}
	return n, nil
}

// Favorite saves a listing for the caller.
func (s *PropertyService) Favorite(ctx context.Context, userID, propertyID string) error {
	n, err := s.newcomer(ctx, userID)
	if err != nil {
		return err
	}
	p, err := s.props.GetByID(ctx, propertyID)
	if err != nil {
		return err
	}
	if p == nil || !p.IsActive {
		return domain.ErrNotFound
	}
	return s.favs.Add(ctx, n.ID, propertyID)
}

// Unfavorite removes a saved listing.
func (s *PropertyService) Unfavorite(ctx context.Context, userID, propertyID string) error {
	n, err := s.newcomer(ctx, userID)
	if err != nil {
		return err
	}
	return s.favs.Remove(ctx, n.ID, propertyID)
}

// Favorites lists the caller's saved listings.
func (s *PropertyService) Favorites(ctx context.Context, userID string) ([]domain.Property, error) {
	n, err := s.newcomer(ctx, userID)
	if err != nil {
		return nil, err
	}
	list, err := s.favs.List(ctx, n.ID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i] = Public(list[i])
	}
	return list, nil
}
