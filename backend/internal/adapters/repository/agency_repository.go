package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"bogoboom-backend/internal/core/domain"
)

type agencyRepository struct{ pool *DB }

// NewAgencyRepository creates the agency repository.
func NewAgencyRepository(pool *DB) domain.AgencyRepository { return &agencyRepository{pool: pool} }

const agencyCols = `a.id::text, a.name, a.city, a.country::text, a.description, a.verified_at IS NOT NULL,
	a.contact_credits::int4, to_char(a.credits_period, 'YYYY-MM-DD'), a.plan, a.plan_active_until`

func scanAgency(row scanner) (*domain.Agency, error) {
	var a domain.Agency
	var plan string
	if err := row.Scan(&a.ID, &a.Name, &a.City, &a.Country, &a.Description, &a.Verified, &a.ContactCredits, &a.CreditsPeriod, &plan, &a.PlanActiveUntil); err != nil {
		return nil, err
	}
	a.Plan = domain.Plan(plan)
	return &a, nil
}

func (r *agencyRepository) GetByUser(ctx context.Context, userID string) (*domain.Agency, error) {
	a, err := scanAgency(r.pool.QueryRow(ctx, `SELECT `+agencyCols+` FROM agencies a JOIN agency_profiles p ON p.agency_id = a.id WHERE p.user_id = $1::text::uuid`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *agencyRepository) GetByID(ctx context.Context, id string) (*domain.Agency, error) {
	a, err := scanAgency(r.pool.QueryRow(ctx, `SELECT `+agencyCols+` FROM agencies a WHERE a.id = $1::text::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *agencyRepository) SetVerified(ctx context.Context, agencyID string, verified bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE agencies SET verified_at = CASE WHEN $2 THEN COALESCE(verified_at, now()) ELSE NULL END WHERE id = $1::text::uuid`, agencyID, verified)
	return err
}

func (r *agencyRepository) SetPlan(ctx context.Context, agencyID string, plan domain.Plan, activeUntil *time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE agencies SET plan = $2, plan_active_until = $3 WHERE id = $1::text::uuid`, agencyID, string(plan), activeUntil)
	return err
}

func (r *agencyRepository) ResetCreditsIfNewMonth(ctx context.Context, agencyID, period string, credits int) error {
	_, err := r.pool.Exec(ctx, `UPDATE agencies SET contact_credits = $3::int4, credits_period = $2::text::date
		WHERE id = $1::text::uuid AND credits_period < $2::text::date`, agencyID, period, credits)
	return err
}

// Unlock inserts the reveal and spends the credit in one transaction. The credit update is
// conditional (contact_credits > 0), so two concurrent unlocks can never spend the same credit.
func (r *agencyRepository) Unlock(ctx context.Context, agencyID, newcomerID, userID string, useCredit bool) (bool, error) {
	already := false
	err := inTx(ctx, r.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO contact_unlocks (agency_id, newcomer_id, unlocked_by, used_credit)
			VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, $4) ON CONFLICT (agency_id, newcomer_id) DO NOTHING`,
			agencyID, newcomerID, userID, useCredit)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			already = true
			return nil
		}
		if useCredit {
			tag, err := tx.Exec(ctx, `UPDATE agencies SET contact_credits = contact_credits - 1 WHERE id = $1::text::uuid AND contact_credits > 0`, agencyID)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return domain.ErrNoCredits // rolls back the insert above
			}
		}
		return nil
	})
	return already, err
}

func (r *agencyRepository) IsUnlocked(ctx context.Context, agencyID, newcomerID string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM contact_unlocks WHERE agency_id = $1::text::uuid AND newcomer_id = $2::text::uuid)`, agencyID, newcomerID).Scan(&ok)
	return ok, err
}

func (r *agencyRepository) UnlockedIDs(ctx context.Context, agencyID string) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT newcomer_id::text FROM contact_unlocks WHERE agency_id = $1::text::uuid`, agencyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
