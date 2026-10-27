package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"bogoboom-backend/internal/core/domain"
)

var ctx = context.Background()

func pf(v float64) *float64 { return &v }
func pi(v int) *int         { return &v }
func pb(v bool) *bool       { return &v }
func ps(v string) *string   { return &v }

var now = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

type env struct {
	m    *memStore
	reg  *RegistrationService
	nc   *NewcomerService
	lead *LeadService
	prop *PropertyService
	pub  *PublicService
}

func newEnv() *env {
	m := newMem()
	clock := Clock(func() time.Time { return now })
	e := &env{m: m}
	e.reg = NewRegistrationService(m).WithClock(clock)
	e.nc = NewNewcomerService(newcomerRepo{m}, m, propRepo{m}, noteRepo{m}, auditRepo{m})
	e.lead = NewLeadService(agencyRepo{m}, newcomerRepo{m}, propRepo{m}, noteRepo{m}, auditRepo{m}).WithClock(clock)
	e.prop = NewPropertyService(agencyRepo{m}, propRepo{m}, newcomerRepo{m}, noteRepo{m}, favRepo{m})
	e.pub = NewPublicService(inqRepo{m}, agencyRepo{m}, auditRepo{m})
	return e
}

func (e *env) agency(t *testing.T, sub string, verified bool) *domain.User {
	t.Helper()
	u, err := e.reg.RegisterAgency(ctx, Identity{Sub: sub, Email: sub + "@agency.co"}, AgencyRegistration{
		Consent: Consent{Accepted: true}, FullName: "Ana Agent", AgencyName: "Casa " + sub, City: "Bogotá", Country: "co"})
	if err != nil {
		t.Fatal(err)
	}
	e.m.agencies[e.m.agencyOfUser[u.ID]].Verified = verified
	return u
}

func (e *env) listing(t *testing.T, agencyUser *domain.User, price float64) *domain.Property {
	t.Helper()
	p, err := e.prop.Create(ctx, agencyUser.ID, PropertyInput{Title: "Apto Chapinero", Price: price, Type: "apartment", Bedrooms: pi(2),
		Furnished: true, City: "Bogota", District: "Chapinero", Address: "Calle 1 # 2-3"})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func (e *env) newcomer(t *testing.T, sub string, share bool) *domain.User {
	t.Helper()
	u, err := e.reg.RegisterNewcomer(ctx, Identity{Sub: sub, Email: sub + "@mail.com"}, NewcomerRegistration{
		Consent: Consent{Accepted: true, ShareWithAgencies: share}, FullName: "Maria Lopez", Phone: "+57 300 1234567",
		CityOfContract: "Bogotá", BudgetMax: pf(3000000), Bedrooms: pi(2), Furnished: pb(true)})
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func isValidation(err error, field string) bool {
	var v *domain.ValidationError
	if !errors.As(err, &v) {
		return false
	}
	_, ok := v.Fields[field]
	return ok
}

// ---------------------------------------------------------------- registration

func TestRegistrationRules(t *testing.T) {
	e := newEnv()
	good := AgencyRegistration{Consent: Consent{Accepted: true}, FullName: "A", AgencyName: "X", City: "Cali", Country: "CO"}

	if _, err := e.reg.RegisterAgency(ctx, Identity{Sub: "s1"}, good); !errors.Is(err, domain.ErrEmailRequired) {
		t.Fatalf("no e-mail in token: %v", err)
	}
	noConsent := good
	noConsent.Accepted = false
	if _, err := e.reg.RegisterAgency(ctx, Identity{Sub: "s1", Email: "a@b.co"}, noConsent); !errors.Is(err, domain.ErrConsentRequired) {
		t.Fatalf("consent: %v", err)
	}
	bad := good
	bad.Country = "Colombia"
	bad.AgencyName = " "
	if _, err := e.reg.RegisterAgency(ctx, Identity{Sub: "s1", Email: "a@b.co"}, bad); !isValidation(err, "country") || !isValidation(err, "agencyName") {
		t.Fatalf("validation: %v", err)
	}
	u, err := e.reg.RegisterAgency(ctx, Identity{Sub: "s1", Email: "a@b.co"}, good)
	if err != nil || u.Role != domain.RoleAgency {
		t.Fatalf("register: %v", err)
	}
	a := e.m.agencies[e.m.agencyOfUser[u.ID]]
	if a.Verified || a.Plan != domain.PlanFree || a.ContactCredits != 5 || a.CreditsPeriod != "2026-10-01" || a.Country != "CO" {
		t.Fatalf("new agency must be unverified/free with 5 credits: %+v", a)
	}
	if _, err := e.reg.RegisterAgency(ctx, Identity{Sub: "s1", Email: "a@b.co"}, good); !errors.Is(err, domain.ErrAlreadyRegistered) {
		t.Fatalf("duplicate: %v", err)
	}
}

func TestNewcomerRegistrationIsOptInAndValidated(t *testing.T) {
	e := newEnv()
	u := e.newcomer(t, "n1", false)
	n, _ := e.nc.Me(ctx, u.ID)
	if n.AllowContact {
		t.Fatal("agencies must not see a newcomer who did not opt in")
	}
	if n.FirstName != "Maria" {
		t.Fatalf("only the first name is public: %q", n.FirstName)
	}
	_, err := e.reg.RegisterNewcomer(ctx, Identity{Sub: "n2", Email: "n2@mail.com"}, NewcomerRegistration{
		Consent: Consent{Accepted: true}, FullName: "Z", CityOfContract: "Bogota", BudgetMin: pf(10), BudgetMax: pf(5), ArrivalDate: ps("31/12/2026"), Phone: "xx"})
	if !isValidation(err, "budgetMax") || !isValidation(err, "arrivalDate") || !isValidation(err, "phone") {
		t.Fatalf("validation: %v", err)
	}
}

// ---------------------------------------------------------------- leads

func TestLeadsAreBlurredUntilUnlocked(t *testing.T) {
	e := newEnv()
	ag := e.agency(t, "ag1", true)
	e.listing(t, ag, 2800000)
	e.newcomer(t, "n1", true)

	leads, err := e.lead.List(ctx, ag.ID)
	if err != nil || len(leads) != 1 {
		t.Fatalf("leads: %v %v", leads, err)
	}
	if leads[0].Contact != nil || leads[0].Unlocked || leads[0].MatchScore < 90 {
		t.Fatalf("lead must be blurred and scored: %+v", leads[0])
	}
	got, err := e.lead.Unlock(ctx, ag.ID, leads[0].ID, "1.2.3.4")
	if err != nil || got.Contact == nil || got.Contact.Email != "n1@mail.com" || got.Contact.Phone == "" {
		t.Fatalf("unlock: %+v %v", got, err)
	}
	if len(e.m.audits) != 1 || e.m.audits[0].Action != "agency.unlock_contact" || e.m.audits[0].IP != "1.2.3.4" {
		t.Fatalf("unlock must be audited: %+v", e.m.audits)
	}
	again, _ := e.lead.List(ctx, ag.ID)
	if !again[0].Unlocked || again[0].Contact == nil {
		t.Fatal("unlocked lead keeps its contact")
	}
}

func TestUnlockGuards(t *testing.T) {
	e := newEnv()
	verified := e.agency(t, "ag1", true)
	unverified := e.agency(t, "ag2", false)
	e.listing(t, verified, 2800000)
	e.listing(t, unverified, 2800000)
	optIn := e.newcomer(t, "n1", true)
	optOut := e.newcomer(t, "n2", false)
	nOpt, _ := e.nc.Me(ctx, optIn.ID)
	nOut, _ := e.nc.Me(ctx, optOut.ID)

	if _, err := e.lead.Unlock(ctx, unverified.ID, nOpt.ID, ""); !errors.Is(err, domain.ErrAgencyNotVerified) {
		t.Fatalf("unverified agency: %v", err)
	}
	if _, err := e.lead.Unlock(ctx, verified.ID, nOut.ID, ""); !errors.Is(err, domain.ErrContactNotAllowed) {
		t.Fatalf("no opt-in: %v", err)
	}
	if _, err := e.lead.Unlock(ctx, verified.ID, "missing", ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown lead: %v", err)
	}
	// An agency without any matching listing cannot harvest contacts.
	noListing := e.agency(t, "ag3", true)
	if _, err := e.lead.Unlock(ctx, noListing.ID, nOpt.ID, ""); !errors.Is(err, domain.ErrNotMatching) {
		t.Fatalf("not matching: %v", err)
	}
	// Other roles have no agency.
	if _, err := e.lead.List(ctx, optIn.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("newcomer is not an agency: %v", err)
	}
	if len(e.m.audits) != 0 {
		t.Fatal("rejected unlocks must not be audited as reveals")
	}
}

func TestFreeCreditsPaidPlanAndMonthlyReset(t *testing.T) {
	e := newEnv()
	ag := e.agency(t, "ag1", true)
	e.listing(t, ag, 2800000)
	var ids []string
	for i := 0; i < 7; i++ {
		u := e.newcomer(t, "n"+string(rune('a'+i)), true)
		n, _ := e.nc.Me(ctx, u.ID)
		ids = append(ids, n.ID)
	}
	for i := 0; i < 5; i++ {
		if _, err := e.lead.Unlock(ctx, ag.ID, ids[i], ""); err != nil {
			t.Fatalf("unlock %d: %v", i, err)
		}
	}
	if _, err := e.lead.Unlock(ctx, ag.ID, ids[5], ""); !errors.Is(err, domain.ErrNoCredits) {
		t.Fatalf("6th unlock on the free plan: %v", err)
	}
	// Re-opening an already unlocked lead is free.
	if _, err := e.lead.Unlock(ctx, ag.ID, ids[0], ""); err != nil {
		t.Fatalf("idempotent unlock: %v", err)
	}
	if got := e.m.agencies[e.m.agencyOfUser[ag.ID]].ContactCredits; got != 0 {
		t.Fatalf("credits: %d", got)
	}

	// A paid plan unlocks without credits...
	until := now.Add(24 * time.Hour)
	_ = e.pub.SetAgencyPlan(ctx, "admin", e.m.agencyOfUser[ag.ID], domain.PlanPro, &until, "")
	if _, err := e.lead.Unlock(ctx, ag.ID, ids[5], ""); err != nil {
		t.Fatalf("pro unlock: %v", err)
	}
	// ...until it expires.
	past := now.Add(-time.Hour)
	_ = e.pub.SetAgencyPlan(ctx, "admin", e.m.agencyOfUser[ag.ID], domain.PlanPro, &past, "")
	if _, err := e.lead.Unlock(ctx, ag.ID, ids[6], ""); !errors.Is(err, domain.ErrNoCredits) {
		t.Fatalf("expired plan: %v", err)
	}
	// Next month the free credits come back by themselves.
	e.lead.WithClock(func() time.Time { return now.AddDate(0, 1, 0) })
	if _, err := e.lead.Unlock(ctx, ag.ID, ids[6], ""); err != nil {
		t.Fatalf("after monthly reset: %v", err)
	}
}

// ---------------------------------------------------------------- HR

func TestHRInviteStaysPrivateUntilTheNewcomerOptsIn(t *testing.T) {
	e := newEnv()
	ag := e.agency(t, "ag1", true)
	e.listing(t, ag, 2800000)
	hr, err := e.reg.RegisterHR(ctx, Identity{Sub: "hr1", Email: "hr@empresa.co"}, HRRegistration{
		Consent: Consent{Accepted: true}, FullName: "Rita HR", CompanyName: "Empresa", Country: "co"})
	if err != nil {
		t.Fatal(err)
	}
	n, err := e.nc.Invite(ctx, hr.ID, "9.9.9.9", Invite{FullName: "Peter Smith", Email: "peter@mail.com", CityOfContract: "Bogota", BudgetMax: pf(3000000)})
	if err != nil || n.AllowContact || n.Stage != domain.StageInvited || n.FirstName != "Peter" {
		t.Fatalf("invite: %+v %v", n, err)
	}
	if leads, _ := e.lead.List(ctx, ag.ID); len(leads) != 0 {
		t.Fatal("an invited newcomer is invisible to agencies until they opt in")
	}
	if list, _ := e.nc.ListForHR(ctx, hr.ID); len(list) != 1 {
		t.Fatalf("HR sees its hires: %v", list)
	}
	// Peter signs up with the invited e-mail: the invitation is claimed, not duplicated.
	if _, err := e.reg.RegisterNewcomer(ctx, Identity{Sub: "peter", Email: "peter@mail.com"}, NewcomerRegistration{
		Consent: Consent{Accepted: true, ShareWithAgencies: true}, FullName: "Peter Smith", CityOfContract: "Bogota"}); err != nil {
		t.Fatal(err)
	}
	if len(e.m.newcomers) != 1 {
		t.Fatalf("claim must not duplicate: %d", len(e.m.newcomers))
	}
	// Non-HR users cannot invite.
	if _, err := e.nc.Invite(ctx, ag.ID, "", Invite{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("agency inviting: %v", err)
	}
	if _, err := e.nc.Invite(ctx, hr.ID, "", Invite{FullName: "x", Email: "bad", CityOfContract: ""}); !isValidation(err, "email") || !isValidation(err, "cityOfContract") {
		t.Fatalf("validation: %v", err)
	}
}

func TestOptInNotifiesMatchingAgencies(t *testing.T) {
	e := newEnv()
	ag := e.agency(t, "ag1", true)
	other := e.agency(t, "ag2", true)
	e.listing(t, ag, 2800000)
	e.listing(t, other, 90000000) // far over budget: no notification
	u := e.newcomer(t, "n1", false)
	if len(e.m.notes) != 0 {
		t.Fatal("no notification before opt-in")
	}
	if _, err := e.nc.Update(ctx, u.ID, "", NewcomerUpdate{AllowContact: pb(true)}); err != nil {
		t.Fatal(err)
	}
	notes, _ := e.lead.Notifications(ctx, ag.ID)
	if len(notes) != 1 || notes[0].MatchScore < domain.NotifyThreshold {
		t.Fatalf("notes: %+v", notes)
	}
	if other, _ := e.lead.Notifications(ctx, other.ID); len(other) != 0 {
		t.Fatalf("non-matching agency notified: %+v", other)
	}
	// Saving again does not duplicate the notification.
	_, _ = e.nc.Update(ctx, u.ID, "", NewcomerUpdate{AllowContact: pb(true)})
	if notes, _ = e.lead.Notifications(ctx, ag.ID); len(notes) != 1 {
		t.Fatal("notifications must be deduplicated")
	}
	if err := e.lead.MarkRead(ctx, ag.ID, notes[0].ID); err != nil {
		t.Fatal(err)
	}
	// Withdrawing consent hides the newcomer from agencies again.
	_, _ = e.nc.Update(ctx, u.ID, "", NewcomerUpdate{AllowContact: pb(false)})
	if leads, _ := e.lead.List(ctx, ag.ID); len(leads) != 0 {
		t.Fatal("withdrawn consent must hide the lead")
	}
}

func TestNewListingNotifiesExistingNewcomers(t *testing.T) {
	e := newEnv()
	e.newcomer(t, "n1", true)
	ag := e.agency(t, "ag1", true)
	e.listing(t, ag, 2800000)
	if notes, _ := e.lead.Notifications(ctx, ag.ID); len(notes) != 1 {
		t.Fatalf("a new listing should notify about waiting newcomers: %v", notes)
	}
}

// ---------------------------------------------------------------- properties

func TestPropertyOwnershipAndPublicView(t *testing.T) {
	e := newEnv()
	a1, a2 := e.agency(t, "ag1", true), e.agency(t, "ag2", true)
	p := e.listing(t, a1, 2800000)

	if _, err := e.prop.Update(ctx, a2.ID, p.ID, PropertyInput{Title: "hijack", Type: "room", City: "Bogota"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("another agency must not edit: %v", err)
	}
	if err := e.prop.Delete(ctx, a2.ID, p.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("another agency must not delete: %v", err)
	}
	if _, err := e.prop.Create(ctx, a1.ID, PropertyInput{Title: "", Price: -1, Type: "castle", City: ""}); !isValidation(err, "title") || !isValidation(err, "price") || !isValidation(err, "type") || !isValidation(err, "city") {
		t.Fatalf("validation: %v", err)
	}
	nu := e.newcomer(t, "n1", false)
	res, err := e.prop.Search(ctx, "bogota", nil)
	if err != nil || len(res) != 1 || res[0].Address != "" {
		t.Fatalf("newcomers must not see street addresses: %+v %v", res, err)
	}
	if _, err := e.prop.Search(ctx, " ", nil); !isValidation(err, "city") {
		t.Fatalf("city is required: %v", err)
	}
	if err := e.prop.Favorite(ctx, nu.ID, p.ID); err != nil {
		t.Fatal(err)
	}
	if favs, _ := e.prop.Favorites(ctx, nu.ID); len(favs) != 1 || favs[0].Address != "" {
		t.Fatalf("favorites: %+v", favs)
	}
	if err := e.prop.Favorite(ctx, a1.ID, p.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("only newcomers favorite: %v", err)
	}
}

// ---------------------------------------------------------------- public forms / admin

func TestPublicForms(t *testing.T) {
	e := newEnv()
	ok := InquiryInput{FullName: "Carlos", Email: "c@x.co", Country: "CO", CollaborationType: "real_estate", Consent: true}
	if err := e.pub.CreatePartnerInquiry(ctx, ok); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*InquiryInput){
		"email":             func(i *InquiryInput) { i.Email = "nope" },
		"collaborationType": func(i *InquiryInput) { i.CollaborationType = "hack" },
		"acceptPrivacy":     func(i *InquiryInput) { i.Consent = false },
		"fullName":          func(i *InquiryInput) { i.FullName = "  " },
	} {
		bad := ok
		mut(&bad)
		if err := e.pub.CreatePartnerInquiry(ctx, bad); !isValidation(err, name) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := e.pub.CreateLead(ctx, LeadInput{Name: "Ana", Email: "a@b.co", Consent: true}); err != nil {
		t.Fatal(err)
	}
	if err := e.pub.CreateLead(ctx, LeadInput{Name: "Ana", Email: "a@b.co"}); !isValidation(err, "acceptPrivacy") {
		t.Fatalf("consent: %v", err)
	}
	if len(e.m.partner) != 1 || len(e.m.leads) != 1 {
		t.Fatal("only valid submissions are stored")
	}
}

func TestAdminActionsAreAuditedAndValidated(t *testing.T) {
	e := newEnv()
	ag := e.agency(t, "ag1", false)
	id := e.m.agencyOfUser[ag.ID]
	if err := e.pub.VerifyAgency(ctx, "admin1", id, true, "5.5.5.5"); err != nil {
		t.Fatal(err)
	}
	if !e.m.agencies[id].Verified || len(e.m.audits) != 1 || e.m.audits[0].UserID != "admin1" {
		t.Fatalf("verify must be audited: %+v", e.m.audits)
	}
	if err := e.pub.VerifyAgency(ctx, "admin1", "nope", true, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown agency: %v", err)
	}
	if err := e.pub.SetAgencyPlan(ctx, "admin1", id, domain.Plan("gold"), nil, ""); !isValidation(err, "plan") {
		t.Fatalf("plan: %v", err)
	}
}

func TestPlanActive(t *testing.T) {
	future, past := now.Add(time.Hour), now.Add(-time.Hour)
	cases := []struct {
		a    domain.Agency
		want bool
	}{
		{domain.Agency{Plan: domain.PlanFree}, false},
		{domain.Agency{Plan: domain.PlanPro}, true},
		{domain.Agency{Plan: domain.PlanPro, PlanActiveUntil: &future}, true},
		{domain.Agency{Plan: domain.PlanEnterprise, PlanActiveUntil: &past}, false},
	}
	for i, c := range cases {
		if PlanActive(&c.a, now) != c.want {
			t.Errorf("case %d", i)
		}
	}
}
