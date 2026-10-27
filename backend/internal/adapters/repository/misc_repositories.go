package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"bogoboom-backend/internal/core/domain"
)

type notificationRepository struct{ pool *DB }

// NewNotificationRepository creates the notification repository.
func NewNotificationRepository(pool *DB) domain.NotificationRepository {
	return &notificationRepository{pool: pool}
}

func (r *notificationRepository) Upsert(ctx context.Context, n *domain.Notification) (bool, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO notifications (agency_id, newcomer_id, match_score, message)
		VALUES ($1::text::uuid, $2::text::uuid, $3::int4, $4)
		ON CONFLICT (agency_id, newcomer_id) DO NOTHING RETURNING id::text, created_at`,
		n.AgencyID, n.NewcomerID, n.MatchScore, n.Message).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // already notified
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *notificationRepository) ListByAgency(ctx context.Context, agencyID string, limit int) ([]domain.Notification, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, agency_id::text, newcomer_id::text, match_score::int4, message, read_at, created_at
		FROM notifications WHERE agency_id = $1::text::uuid ORDER BY created_at DESC LIMIT $2`, agencyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.AgencyID, &n.NewcomerID, &n.MatchScore, &n.Message, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *notificationRepository) MarkRead(ctx context.Context, agencyID, id string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE notifications SET read_at = COALESCE(read_at, now())
		WHERE id = $1::text::uuid AND agency_id = $2::text::uuid`, id, agencyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type inquiryRepository struct{ pool *DB }

// NewInquiryRepository creates the repository of public form submissions.
func NewInquiryRepository(pool *DB) domain.InquiryRepository { return &inquiryRepository{pool: pool} }

func (r *inquiryRepository) CreatePartnerInquiry(ctx context.Context, q *domain.PartnerInquiry) error {
	return r.pool.QueryRow(ctx, `INSERT INTO partner_inquiries (full_name, email, phone, country, city, collaboration_type)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id::text, created_at`,
		q.FullName, q.Email, q.Phone, q.Country, q.City, q.CollaborationType).Scan(&q.ID, &q.CreatedAt)
}

func (r *inquiryRepository) ListPartnerInquiries(ctx context.Context, limit int) ([]domain.PartnerInquiry, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, full_name, email, phone, country, city, collaboration_type, created_at
		FROM partner_inquiries ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PartnerInquiry{}
	for rows.Next() {
		var q domain.PartnerInquiry
		if err := rows.Scan(&q.ID, &q.FullName, &q.Email, &q.Phone, &q.Country, &q.City, &q.CollaborationType, &q.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *inquiryRepository) CreateLandingLead(ctx context.Context, l *domain.LandingLead) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO landing_leads (name, email, phone, city, country, interest) VALUES ($1, $2, $3, $4, $5, $6)`,
		l.Name, l.Email, l.Phone, l.City, l.Country, l.Interest)
	return err
}

type auditRepository struct{ pool *DB }

// NewAuditRepository creates the audit log repository.
func NewAuditRepository(pool *DB) domain.AuditRepository { return &auditRepository{pool: pool} }

func (r *auditRepository) Log(ctx context.Context, e domain.AuditEntry) error {
	changes, err := json.Marshal(e.Changes)
	if err != nil || e.Changes == nil {
		changes = []byte("{}")
	}
	var userID *string
	if e.UserID != "" {
		userID = &e.UserID
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO audit_logs (user_id, action, table_name, record_id, changes, ip)
		VALUES ($1::text::uuid, $2, $3, $4, $5::text::jsonb, $6)`, userID, e.Action, e.TableName, e.RecordID, string(changes), e.IP)
	return err
}
