package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"bogoboom-backend/internal/core/domain"
)

type newcomerRepository struct{ pool *DB }

// NewNewcomerRepository creates the newcomer repository.
func NewNewcomerRepository(pool *DB) domain.NewcomerRepository {
	return &newcomerRepository{pool: pool}
}

const newcomerCols = `n.id::text, n.user_id::text, n.company_id::text, n.created_by::text, n.first_name, n.city_of_contract,
	to_char(n.arrival_date, 'YYYY-MM-DD'), n.budget_min::float8, n.budget_max::float8, n.currency::text, n.bedrooms::int4,
	n.furnished, n.preferred_districts, n.stage::text, n.allow_contact, n.created_at`

func scanNewcomer(row scanner) (*domain.Newcomer, error) {
	var n domain.Newcomer
	var stage string
	if err := row.Scan(&n.ID, &n.UserID, &n.CompanyID, &n.CreatedBy, &n.FirstName, &n.CityOfContract, &n.ArrivalDate,
		&n.BudgetMin, &n.BudgetMax, &n.Currency, &n.Bedrooms, &n.Furnished, &n.PreferredDistricts, &stage, &n.AllowContact, &n.CreatedAt); err != nil {
		return nil, err
	}
	n.Stage = domain.RelocationStage(stage)
	if n.PreferredDistricts == nil {
		n.PreferredDistricts = []string{}
	}
	return &n, nil
}

func (r *newcomerRepository) one(ctx context.Context, where string, arg any) (*domain.Newcomer, error) {
	n, err := scanNewcomer(r.pool.QueryRow(ctx, `SELECT `+newcomerCols+` FROM newcomers n WHERE n.deleted_at IS NULL AND `+where, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return n, err
}

func (r *newcomerRepository) many(ctx context.Context, sql string, args ...any) ([]domain.Newcomer, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Newcomer{}
	for rows.Next() {
		n, err := scanNewcomer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

func (r *newcomerRepository) CreateInvited(ctx context.Context, n *domain.Newcomer, c *domain.NewcomerContact) error {
	return inTx(ctx, r.pool, func(tx pgx.Tx) error {
		if err := insertNewcomer(ctx, tx, n); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO newcomer_contacts (newcomer_id, full_name, email, phone) VALUES ($1::text::uuid, $2, $3, $4)`,
			n.ID, c.FullName, c.Email, c.Phone)
		return err
	})
}

func (r *newcomerRepository) GetByID(ctx context.Context, id string) (*domain.Newcomer, error) {
	return r.one(ctx, `n.id = $1::text::uuid`, id)
}

func (r *newcomerRepository) GetByUserID(ctx context.Context, userID string) (*domain.Newcomer, error) {
	return r.one(ctx, `n.user_id = $1::text::uuid`, userID)
}

func (r *newcomerRepository) ListByCompany(ctx context.Context, companyID string) ([]domain.Newcomer, error) {
	return r.many(ctx, `SELECT `+newcomerCols+` FROM newcomers n WHERE n.deleted_at IS NULL AND n.company_id = $1::text::uuid ORDER BY n.created_at DESC LIMIT 500`, companyID)
}

func (r *newcomerRepository) ListOpen(ctx context.Context, limit int) ([]domain.Newcomer, error) {
	return r.many(ctx, `SELECT `+newcomerCols+` FROM newcomers n
		WHERE n.deleted_at IS NULL AND n.allow_contact AND n.stage NOT IN ('cancelled','moved_in')
		ORDER BY n.created_at DESC LIMIT $1`, limit)
}

func (r *newcomerRepository) Update(ctx context.Context, n *domain.Newcomer) error {
	tag, err := r.pool.Exec(ctx, `UPDATE newcomers SET city_of_contract = $2, city_key = $3, arrival_date = $4::text::date,
		budget_min = $5::float8, budget_max = $6::float8, bedrooms = $7::int4, furnished = $8, preferred_districts = $9,
		stage = $10::text::relocation_stage, allow_contact = $11, updated_at = now()
		WHERE id = $1::text::uuid AND deleted_at IS NULL`,
		n.ID, n.CityOfContract, domain.Normalize(n.CityOfContract), n.ArrivalDate, n.BudgetMin, n.BudgetMax, n.Bedrooms,
		n.Furnished, n.PreferredDistricts, string(n.Stage), n.AllowContact)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *newcomerRepository) GetContact(ctx context.Context, newcomerID string) (*domain.NewcomerContact, error) {
	var c domain.NewcomerContact
	err := r.pool.QueryRow(ctx, `SELECT full_name, email, phone FROM newcomer_contacts WHERE newcomer_id = $1::text::uuid`, newcomerID).
		Scan(&c.FullName, &c.Email, &c.Phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

type companyRepository struct{ pool *DB }

// NewCompanyRepository creates the company repository.
func NewCompanyRepository(pool *DB) domain.CompanyRepository { return &companyRepository{pool: pool} }

func (r *companyRepository) GetByHRUser(ctx context.Context, userID string) (*domain.Company, error) {
	var c domain.Company
	err := r.pool.QueryRow(ctx, `SELECT c.id::text, c.name, c.industry, c.city, c.country::text, c.allow_agency_contact
		FROM hr_profiles h JOIN companies c ON c.id = h.company_id WHERE h.user_id = $1::text::uuid`, userID).
		Scan(&c.ID, &c.Name, &c.Industry, &c.City, &c.Country, &c.AllowAgencyContact)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}
