package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"bogoboom-backend/internal/core/domain"
)

// memStore is one in-memory database implementing every repository port.
type memStore struct {
	mu           sync.Mutex
	seq          int
	users        map[string]*domain.User
	newcomers    map[string]*domain.Newcomer
	contacts     map[string]*domain.NewcomerContact
	companies    map[string]*domain.Company
	hrCompany    map[string]string // user -> company
	agencies     map[string]*domain.Agency
	agencyOfUser map[string]string
	props        map[string]*domain.Property
	unlocks      map[string]bool // agency|newcomer
	notes        map[string]*domain.Notification
	favs         map[string]bool
	partner      []domain.PartnerInquiry
	leads        []domain.LandingLead
	audits       []domain.AuditEntry
	resetCalls   int
}

func newMem() *memStore {
	return &memStore{users: map[string]*domain.User{}, newcomers: map[string]*domain.Newcomer{}, contacts: map[string]*domain.NewcomerContact{},
		companies: map[string]*domain.Company{}, hrCompany: map[string]string{}, agencies: map[string]*domain.Agency{},
		agencyOfUser: map[string]string{}, props: map[string]*domain.Property{}, unlocks: map[string]bool{},
		notes: map[string]*domain.Notification{}, favs: map[string]bool{}}
}

func (m *memStore) id(prefix string) string { m.seq++; return fmt.Sprintf("%s-%d", prefix, m.seq) }

// ---- UserRepository
func (m *memStore) GetByAuth0Sub(_ context.Context, sub string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Auth0Sub == sub {
			c := *u
			return &c, nil
		}
	}
	return nil, nil
}
func (m *memStore) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		c := *u
		return &c, nil
	}
	return nil, nil
}
func (m *memStore) CreateNewcomer(_ context.Context, u *domain.User, n *domain.Newcomer, c *domain.NewcomerContact) error {
	u.ID = m.id("u")
	m.users[u.ID] = u
	for _, ex := range m.newcomers { // claim an invitation made by HR for the same e-mail
		if ex.UserID == nil && strings.EqualFold(m.contacts[ex.ID].Email, u.Email) {
			ex.UserID = &u.ID
			ex.AllowContact = n.AllowContact
			ex.Stage = domain.StageProfileDone
			return nil
		}
	}
	n.ID = m.id("n")
	n.UserID = &u.ID
	m.newcomers[n.ID], m.contacts[n.ID] = n, c
	return nil
}
func (m *memStore) CreateHR(_ context.Context, u *domain.User, co *domain.Company, _, _ string) error {
	u.ID, co.ID = m.id("u"), m.id("co")
	m.users[u.ID], m.companies[co.ID], m.hrCompany[u.ID] = u, co, co.ID
	return nil
}
func (m *memStore) CreateAgency(_ context.Context, u *domain.User, a *domain.Agency, _ string) error {
	u.ID, a.ID = m.id("u"), m.id("ag")
	m.users[u.ID], m.agencies[a.ID], m.agencyOfUser[u.ID] = u, a, a.ID
	return nil
}

// ---- NewcomerRepository
func (m *memStore) CreateInvited(_ context.Context, n *domain.Newcomer, c *domain.NewcomerContact) error {
	n.ID = m.id("n")
	m.newcomers[n.ID], m.contacts[n.ID] = n, c
	return nil
}
func (m *memStore) GetByIDNewcomer(id string) *domain.Newcomer { return m.newcomers[id] }
func (m *memStore) GetByUserID(_ context.Context, uid string) (*domain.Newcomer, error) {
	for _, n := range m.newcomers {
		if n.UserID != nil && *n.UserID == uid {
			c := *n
			return &c, nil
		}
	}
	return nil, nil
}
func (m *memStore) ListByCompany(_ context.Context, cid string) ([]domain.Newcomer, error) {
	var out []domain.Newcomer
	for _, n := range m.newcomers {
		if n.CompanyID != nil && *n.CompanyID == cid {
			out = append(out, *n)
		}
	}
	return out, nil
}
func (m *memStore) ListOpen(_ context.Context, _ int) ([]domain.Newcomer, error) {
	var out []domain.Newcomer
	for _, n := range m.newcomers {
		if n.AllowContact && n.Stage != domain.StageCancelled && n.Stage != domain.StageMovedIn {
			out = append(out, *n)
		}
	}
	return out, nil
}
func (m *memStore) Update(_ context.Context, n *domain.Newcomer) error {
	c := *n
	m.newcomers[n.ID] = &c
	return nil
}
func (m *memStore) GetContact(_ context.Context, id string) (*domain.NewcomerContact, error) {
	c, ok := m.contacts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cc := *c
	return &cc, nil
}

// newcomerRepo adapts memStore.GetByID (name clash between user and newcomer ports).
type newcomerRepo struct{ *memStore }

func (r newcomerRepo) GetByID(_ context.Context, id string) (*domain.Newcomer, error) {
	if n, ok := r.newcomers[id]; ok {
		c := *n
		return &c, nil
	}
	return nil, nil
}

// ---- CompanyRepository
func (m *memStore) GetByHRUser(_ context.Context, uid string) (*domain.Company, error) {
	if cid, ok := m.hrCompany[uid]; ok {
		c := *m.companies[cid]
		return &c, nil
	}
	return nil, nil
}

// ---- AgencyRepository
type agencyRepo struct{ *memStore }

func (r agencyRepo) GetByUser(_ context.Context, uid string) (*domain.Agency, error) {
	if id, ok := r.agencyOfUser[uid]; ok {
		c := *r.agencies[id]
		return &c, nil
	}
	return nil, nil
}
func (r agencyRepo) GetByID(_ context.Context, id string) (*domain.Agency, error) {
	if a, ok := r.agencies[id]; ok {
		c := *a
		return &c, nil
	}
	return nil, nil
}
func (r agencyRepo) SetVerified(_ context.Context, id string, v bool) error {
	r.agencies[id].Verified = v
	return nil
}
func (r agencyRepo) SetPlan(_ context.Context, id string, p domain.Plan, until *time.Time) error {
	r.agencies[id].Plan, r.agencies[id].PlanActiveUntil = p, until
	return nil
}
func (r agencyRepo) ResetCreditsIfNewMonth(_ context.Context, id, period string, credits int) error {
	r.resetCalls++
	a := r.agencies[id]
	if a.CreditsPeriod < period {
		a.CreditsPeriod, a.ContactCredits = period, credits
	}
	return nil
}
func (r agencyRepo) Unlock(_ context.Context, agencyID, newcomerID, _ string, useCredit bool) (bool, error) {
	k := agencyID + "|" + newcomerID
	if r.unlocks[k] {
		return true, nil
	}
	a := r.agencies[agencyID]
	if useCredit {
		if a.ContactCredits <= 0 {
			return false, domain.ErrNoCredits
		}
		a.ContactCredits--
	}
	r.unlocks[k] = true
	return false, nil
}
func (r agencyRepo) IsUnlocked(_ context.Context, a, n string) (bool, error) {
	return r.unlocks[a+"|"+n], nil
}
func (r agencyRepo) UnlockedIDs(_ context.Context, a string) (map[string]bool, error) {
	out := map[string]bool{}
	for k := range r.unlocks {
		if strings.HasPrefix(k, a+"|") {
			out[strings.TrimPrefix(k, a+"|")] = true
		}
	}
	return out, nil
}

// ---- PropertyRepository
type propRepo struct{ *memStore }

func (r propRepo) Create(_ context.Context, p *domain.Property) error {
	p.ID = r.id("p")
	c := *p
	r.props[p.ID] = &c
	return nil
}
func (r propRepo) Update(_ context.Context, p *domain.Property) error {
	if ex, ok := r.props[p.ID]; !ok || ex.AgencyID != p.AgencyID {
		return domain.ErrNotFound
	}
	c := *p
	r.props[p.ID] = &c
	return nil
}
func (r propRepo) Delete(_ context.Context, agencyID, id string) error {
	if ex, ok := r.props[id]; !ok || ex.AgencyID != agencyID {
		return domain.ErrNotFound
	}
	delete(r.props, id)
	return nil
}
func (r propRepo) GetByID(_ context.Context, id string) (*domain.Property, error) {
	if p, ok := r.props[id]; ok {
		c := *p
		return &c, nil
	}
	return nil, nil
}
func (r propRepo) ListByAgency(_ context.Context, agencyID string) ([]domain.Property, error) {
	var out []domain.Property
	for _, p := range r.props {
		if p.AgencyID == agencyID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (r propRepo) Search(_ context.Context, city string, max *float64, _ int) ([]domain.Property, error) {
	var out []domain.Property
	for _, p := range r.props {
		if p.IsActive && domain.Normalize(p.City) == domain.Normalize(city) && (max == nil || p.Price <= *max) {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (r propRepo) ActiveByCity(ctx context.Context, city string) ([]domain.Property, error) {
	return r.Search(ctx, city, nil, 0)
}

// ---- FavoriteRepository
type favRepo struct{ *memStore }

func (r favRepo) Add(_ context.Context, n, p string) error { r.favs[n+"|"+p] = true; return nil }
func (r favRepo) Remove(_ context.Context, n, p string) error {
	delete(r.favs, n+"|"+p)
	return nil
}
func (r favRepo) List(_ context.Context, n string) ([]domain.Property, error) {
	var out []domain.Property
	for k := range r.favs {
		if strings.HasPrefix(k, n+"|") {
			out = append(out, *r.props[strings.TrimPrefix(k, n+"|")])
		}
	}
	return out, nil
}

// ---- NotificationRepository
type noteRepo struct{ *memStore }

func (r noteRepo) Upsert(_ context.Context, n *domain.Notification) (bool, error) {
	k := n.AgencyID + "|" + n.NewcomerID
	if _, ok := r.notes[k]; ok {
		return false, nil
	}
	n.ID = r.id("no")
	c := *n
	r.notes[k] = &c
	return true, nil
}
func (r noteRepo) ListByAgency(_ context.Context, agencyID string, _ int) ([]domain.Notification, error) {
	var out []domain.Notification
	for _, n := range r.notes {
		if n.AgencyID == agencyID {
			out = append(out, *n)
		}
	}
	return out, nil
}
func (r noteRepo) MarkRead(_ context.Context, agencyID, id string) error {
	for _, n := range r.notes {
		if n.ID == id && n.AgencyID == agencyID {
			now := time.Now()
			n.ReadAt = &now
			return nil
		}
	}
	return domain.ErrNotFound
}

// ---- InquiryRepository / AuditRepository
type inqRepo struct{ *memStore }

func (r inqRepo) CreatePartnerInquiry(_ context.Context, q *domain.PartnerInquiry) error {
	r.partner = append(r.partner, *q)
	return nil
}
func (r inqRepo) ListPartnerInquiries(_ context.Context, _ int) ([]domain.PartnerInquiry, error) {
	return r.partner, nil
}
func (r inqRepo) CreateLandingLead(_ context.Context, l *domain.LandingLead) error {
	r.leads = append(r.leads, *l)
	return nil
}

type auditRepo struct{ *memStore }

func (r auditRepo) Log(_ context.Context, e domain.AuditEntry) error {
	r.audits = append(r.audits, e)
	return nil
}
