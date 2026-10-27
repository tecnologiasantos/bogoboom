package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"bogoboom-backend/internal/core/domain"
)

// PolicyVersion is stored with every consent so a later policy change can ask again.
const PolicyVersion = "2026-10"

type userRepository struct{ pool *DB }

// NewUserRepository creates the user repository.
func NewUserRepository(pool *DB) domain.UserRepository { return &userRepository{pool: pool} }

const userCols = `id::text, auth0_sub, email, role::text, locale, is_active, created_at`

func scanUser(row scanner) (*domain.User, error) {
	var u domain.User
	var role string
	if err := row.Scan(&u.ID, &u.Auth0Sub, &u.Email, &role, &u.Locale, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	u.Role = domain.UserRole(role)
	return &u, nil
}

func (r *userRepository) GetByAuth0Sub(ctx context.Context, sub string) (*domain.User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE auth0_sub = $1`, sub))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1::text::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func insertUser(ctx context.Context, tx querier, u *domain.User, shareWithAgencies *bool) error {
	err := tx.QueryRow(ctx,
		`INSERT INTO users (auth0_sub, email, role, locale) VALUES ($1, $2, $3::text::user_role, $4)
		 RETURNING id::text, created_at`, u.Auth0Sub, u.Email, string(u.Role), u.Locale).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyRegistered
		}
		return err
	}
	for _, c := range []struct {
		purpose string
		granted bool
	}{{"terms", true}, {"privacy", true}} {
		if _, err := tx.Exec(ctx, `INSERT INTO consents (user_id, purpose, granted, policy_version) VALUES ($1::text::uuid, $2, $3, $4)`,
			u.ID, c.purpose, c.granted, PolicyVersion); err != nil {
			return err
		}
	}
	if shareWithAgencies != nil {
		if _, err := tx.Exec(ctx, `INSERT INTO consents (user_id, purpose, granted, policy_version) VALUES ($1::text::uuid, 'share_with_agencies', $2, $3)`,
			u.ID, *shareWithAgencies, PolicyVersion); err != nil {
			return err
		}
	}
	u.IsActive = true
	return nil
}

const newcomerInsert = `INSERT INTO newcomers (user_id, company_id, created_by, first_name, city_of_contract, city_key, arrival_date,
	budget_min, budget_max, currency, bedrooms, furnished, preferred_districts, stage, allow_contact)
	VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, $4, $5, $6, $7::text::date, $8::float8, $9::float8, $10, $11::int4, $12, $13, $14::text::relocation_stage, $15)
	RETURNING id::text, created_at`

func insertNewcomer(ctx context.Context, tx querier, n *domain.Newcomer) error {
	if n.PreferredDistricts == nil {
		n.PreferredDistricts = []string{}
	}
	return tx.QueryRow(ctx, newcomerInsert, n.UserID, n.CompanyID, n.CreatedBy, n.FirstName, n.CityOfContract,
		domain.Normalize(n.CityOfContract), n.ArrivalDate, n.BudgetMin, n.BudgetMax, n.Currency, n.Bedrooms, n.Furnished,
		n.PreferredDistricts, string(n.Stage), n.AllowContact).Scan(&n.ID, &n.CreatedAt)
}

func (r *userRepository) CreateNewcomer(ctx context.Context, u *domain.User, n *domain.Newcomer, c *domain.NewcomerContact) error {
	return inTx(ctx, r.pool, func(tx pgx.Tx) error {
		share := n.AllowContact
		if err := insertUser(ctx, tx, u, &share); err != nil {
			return err
		}
		// If HR already invited this e-mail, claim that profile instead of creating a second one.
		var invited string
		err := tx.QueryRow(ctx,
			`SELECT n.id::text FROM newcomers n JOIN newcomer_contacts c ON c.newcomer_id = n.id
			  WHERE n.user_id IS NULL AND lower(c.email) = lower($1) AND n.deleted_at IS NULL
			  ORDER BY n.created_at LIMIT 1 FOR UPDATE OF n`, u.Email).Scan(&invited)
		switch {
		case err == nil:
			n.ID = invited
			_, err = tx.Exec(ctx,
				`UPDATE newcomers SET user_id = $2::text::uuid, allow_contact = $3, stage = 'profile_done',
				        first_name = $4, city_of_contract = $5, city_key = $6, arrival_date = COALESCE($7::text::date, arrival_date),
				        budget_min = COALESCE($8::float8, budget_min), budget_max = COALESCE($9::float8, budget_max),
				        bedrooms = COALESCE($10::int4, bedrooms), furnished = COALESCE($11, furnished), updated_at = now()
				  WHERE id = $1::text::uuid`,
				invited, u.ID, n.AllowContact, n.FirstName, n.CityOfContract, domain.Normalize(n.CityOfContract),
				n.ArrivalDate, n.BudgetMin, n.BudgetMax, n.Bedrooms, n.Furnished)
			if err != nil {
				return err
			}
			n.UserID = &u.ID
			return nil
		case errors.Is(err, pgx.ErrNoRows):
			n.UserID = &u.ID
			if err := insertNewcomer(ctx, tx, n); err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO newcomer_contacts (newcomer_id, full_name, email, phone) VALUES ($1::text::uuid, $2, $3, $4)`,
				n.ID, c.FullName, c.Email, c.Phone)
			return err
		default:
			return err
		}
	})
}

func (r *userRepository) CreateHR(ctx context.Context, u *domain.User, co *domain.Company, fullName, phone string) error {
	return inTx(ctx, r.pool, func(tx pgx.Tx) error {
		if err := insertUser(ctx, tx, u, nil); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO companies (name, industry, city, country, allow_agency_contact) VALUES ($1, $2, $3, $4, $5) RETURNING id::text`,
			co.Name, co.Industry, co.City, co.Country, co.AllowAgencyContact).Scan(&co.ID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO hr_profiles (user_id, company_id, full_name, phone) VALUES ($1::text::uuid, $2::text::uuid, $3, $4)`,
			u.ID, co.ID, fullName, phone)
		return err
	})
}

func (r *userRepository) CreateAgency(ctx context.Context, u *domain.User, a *domain.Agency, fullName string) error {
	return inTx(ctx, r.pool, func(tx pgx.Tx) error {
		if err := insertUser(ctx, tx, u, nil); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO agencies (name, city, country, description, contact_credits, credits_period)
			 VALUES ($1, $2, $3, $4, $5, $6::text::date) RETURNING id::text`,
			a.Name, a.City, a.Country, a.Description, a.ContactCredits, a.CreditsPeriod).Scan(&a.ID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO agency_profiles (user_id, agency_id, full_name) VALUES ($1::text::uuid, $2::text::uuid, $3)`,
			u.ID, a.ID, fullName)
		return err
	})
}
