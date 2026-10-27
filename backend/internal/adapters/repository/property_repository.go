package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"bogoboom-backend/internal/core/domain"
)

type propertyRepository struct{ pool *DB }

// NewPropertyRepository creates the property repository.
func NewPropertyRepository(pool *DB) domain.PropertyRepository {
	return &propertyRepository{pool: pool}
}

const propertyCols = `p.id::text, p.agency_id::text, p.title, p.description, p.price::float8, p.currency::text, p.type::text,
	p.bedrooms::int4, p.furnished, p.city, p.district, p.address, to_char(p.available_from, 'YYYY-MM-DD'),
	p.verified_at IS NOT NULL, p.is_active, p.created_at`

func scanProperty(row scanner) (*domain.Property, error) {
	var p domain.Property
	var typ string
	if err := row.Scan(&p.ID, &p.AgencyID, &p.Title, &p.Description, &p.Price, &p.Currency, &typ, &p.Bedrooms, &p.Furnished,
		&p.City, &p.District, &p.Address, &p.AvailableFrom, &p.Verified, &p.IsActive, &p.CreatedAt); err != nil {
		return nil, err
	}
	p.Type = domain.PropertyType(typ)
	return &p, nil
}

func (r *propertyRepository) list(ctx context.Context, sql string, args ...any) ([]domain.Property, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Property{}
	for rows.Next() {
		p, err := scanProperty(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *propertyRepository) Create(ctx context.Context, p *domain.Property) error {
	return r.pool.QueryRow(ctx, `INSERT INTO properties (agency_id, title, description, price, currency, type, bedrooms, furnished,
		city, city_key, district, address, available_from, is_active)
		VALUES ($1::text::uuid, $2, $3, $4::float8, $5, $6::text::property_type, $7::int4, $8, $9, $10, $11, $12, $13::text::date, $14)
		RETURNING id::text, created_at`,
		p.AgencyID, p.Title, p.Description, p.Price, p.Currency, string(p.Type), p.Bedrooms, p.Furnished,
		p.City, domain.Normalize(p.City), p.District, p.Address, p.AvailableFrom, p.IsActive).Scan(&p.ID, &p.CreatedAt)
}

func (r *propertyRepository) Update(ctx context.Context, p *domain.Property) error {
	tag, err := r.pool.Exec(ctx, `UPDATE properties SET title = $3, description = $4, price = $5::float8, currency = $6,
		type = $7::text::property_type, bedrooms = $8::int4, furnished = $9, city = $10, city_key = $11, district = $12,
		address = $13, available_from = $14::text::date, is_active = $15, updated_at = now()
		WHERE id = $1::text::uuid AND agency_id = $2::text::uuid`,
		p.ID, p.AgencyID, p.Title, p.Description, p.Price, p.Currency, string(p.Type), p.Bedrooms, p.Furnished,
		p.City, domain.Normalize(p.City), p.District, p.Address, p.AvailableFrom, p.IsActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *propertyRepository) Delete(ctx context.Context, agencyID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM properties WHERE id = $1::text::uuid AND agency_id = $2::text::uuid`, id, agencyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *propertyRepository) GetByID(ctx context.Context, id string) (*domain.Property, error) {
	p, err := scanProperty(r.pool.QueryRow(ctx, `SELECT `+propertyCols+` FROM properties p WHERE p.id = $1::text::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *propertyRepository) ListByAgency(ctx context.Context, agencyID string) ([]domain.Property, error) {
	return r.list(ctx, `SELECT `+propertyCols+` FROM properties p WHERE p.agency_id = $1::text::uuid ORDER BY p.created_at DESC LIMIT 500`, agencyID)
}

func (r *propertyRepository) Search(ctx context.Context, city string, maxPrice *float64, limit int) ([]domain.Property, error) {
	return r.list(ctx, `SELECT `+propertyCols+` FROM properties p
		WHERE p.is_active AND p.city_key = $1 AND ($2::float8 IS NULL OR p.price <= $2::float8)
		ORDER BY p.price ASC LIMIT $3`, domain.Normalize(city), maxPrice, limit)
}

func (r *propertyRepository) ActiveByCity(ctx context.Context, city string) ([]domain.Property, error) {
	return r.list(ctx, `SELECT `+propertyCols+` FROM properties p WHERE p.is_active AND p.city_key = $1 LIMIT 5000`, domain.Normalize(city))
}

type favoriteRepository struct{ pool *DB }

// NewFavoriteRepository creates the favorites repository.
func NewFavoriteRepository(pool *DB) domain.FavoriteRepository {
	return &favoriteRepository{pool: pool}
}

func (r *favoriteRepository) Add(ctx context.Context, newcomerID, propertyID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO favorites (newcomer_id, property_id) VALUES ($1::text::uuid, $2::text::uuid) ON CONFLICT DO NOTHING`, newcomerID, propertyID)
	return err
}

func (r *favoriteRepository) Remove(ctx context.Context, newcomerID, propertyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM favorites WHERE newcomer_id = $1::text::uuid AND property_id = $2::text::uuid`, newcomerID, propertyID)
	return err
}

func (r *favoriteRepository) List(ctx context.Context, newcomerID string) ([]domain.Property, error) {
	pr := &propertyRepository{pool: r.pool}
	return pr.list(ctx, `SELECT `+propertyCols+` FROM favorites f JOIN properties p ON p.id = f.property_id
		WHERE f.newcomer_id = $1::text::uuid AND p.is_active ORDER BY f.created_at DESC LIMIT 200`, newcomerID)
}
