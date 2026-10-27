package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"bogoboom-backend/internal/adapters/repository"
	"bogoboom-backend/internal/core/domain"
	"bogoboom-backend/internal/core/service"
)

// Runs the real services on a real PostgreSQL. Set BOGOBOOM_TEST_DB to a database where
// `go run ./cmd/migrate` has been applied; CI does. Without it the test is skipped.
func TestFullFlowOnPostgres(t *testing.T) {
	dsn := os.Getenv("BOGOBOOM_TEST_DB")
	if dsn == "" {
		t.Skip("BOGOBOOM_TEST_DB not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	run := fmt.Sprintf("%d", time.Now().UnixNano())
	city := "Testópolis " + run // accents on purpose: matching must fold them
	cityPlain := "Testopolis " + run

	users := repository.NewUserRepository(pool)
	newcomers := repository.NewNewcomerRepository(pool)
	companies := repository.NewCompanyRepository(pool)
	agencies := repository.NewAgencyRepository(pool)
	props := repository.NewPropertyRepository(pool)
	favs := repository.NewFavoriteRepository(pool)
	notes := repository.NewNotificationRepository(pool)
	inq := repository.NewInquiryRepository(pool)
	audit := repository.NewAuditRepository(pool)

	reg := service.NewRegistrationService(users)
	nsvc := service.NewNewcomerService(newcomers, companies, props, notes, audit)
	lsvc := service.NewLeadService(agencies, newcomers, props, notes, audit)
	psvc := service.NewPropertyService(agencies, props, newcomers, notes, favs)
	pub := service.NewPublicService(inq, agencies, audit)

	f := func(v float64) *float64 { return &v }
	i := func(v int) *int { return &v }
	b := func(v bool) *bool { return &v }
	consent := service.Consent{Accepted: true, ShareWithAgencies: true}

	// agency + listing
	agUser, err := reg.RegisterAgency(ctx, service.Identity{Sub: "auth0|ag" + run, Email: "ag" + run + "@x.co"},
		service.AgencyRegistration{Consent: service.Consent{Accepted: true}, FullName: "Ana", AgencyName: "Casa " + run, City: city, Country: "co"})
	if err != nil {
		t.Fatalf("register agency: %v", err)
	}
	if _, err := reg.RegisterAgency(ctx, service.Identity{Sub: "auth0|ag" + run, Email: "other" + run + "@x.co"},
		service.AgencyRegistration{Consent: service.Consent{Accepted: true}, FullName: "Ana", AgencyName: "Dup", City: city, Country: "co"}); !errors.Is(err, domain.ErrAlreadyRegistered) {
		t.Fatalf("same subject twice: %v", err)
	}
	prop, err := psvc.Create(ctx, agUser.ID, service.PropertyInput{Title: "Apto", Price: 2800000, Type: "apartment", Bedrooms: i(2), Furnished: true,
		City: city, District: "Chapinero", Address: "Calle 1"})
	if err != nil {
		t.Fatalf("create property: %v", err)
	}

	// HR invites a hire, who then signs up with the same e-mail and opts in
	hr, err := reg.RegisterHR(ctx, service.Identity{Sub: "auth0|hr" + run, Email: "hr" + run + "@x.co"},
		service.HRRegistration{Consent: service.Consent{Accepted: true}, FullName: "Rita", CompanyName: "Emp " + run, Country: "co"})
	if err != nil {
		t.Fatalf("register hr: %v", err)
	}
	inv, err := nsvc.Invite(ctx, hr.ID, "1.1.1.1", service.Invite{FullName: "Peter Smith", Email: "peter" + run + "@x.co", Phone: "+57 300 1234567",
		CityOfContract: cityPlain, BudgetMax: f(3000000), Bedrooms: i(2), Furnished: b(true)})
	if err != nil || inv.AllowContact {
		t.Fatalf("invite: %+v %v", inv, err)
	}
	if hrList, _ := nsvc.ListForHR(ctx, hr.ID); len(hrList) != 1 {
		t.Fatalf("hr list: %v", hrList)
	}
	if leads, err := lsvc.List(ctx, agUser.ID); err != nil || len(leads) != 0 {
		t.Fatalf("invited newcomer must be invisible: %v %v", leads, err)
	}
	peter, err := reg.RegisterNewcomer(ctx, service.Identity{Sub: "auth0|peter" + run, Email: "PETER" + run + "@x.co"},
		service.NewcomerRegistration{Consent: consent, FullName: "Peter Smith", CityOfContract: city})
	if err != nil {
		t.Fatalf("register newcomer (claim): %v", err)
	}
	me, err := nsvc.Me(ctx, peter.ID)
	if err != nil || me.ID != inv.ID || !me.AllowContact || me.BudgetMax == nil || *me.BudgetMax != 3000000 {
		t.Fatalf("claim must reuse the invited profile and keep HR's data: %+v %v", me, err)
	}
	if _, err := nsvc.Update(ctx, peter.ID, "2.2.2.2", service.NewcomerUpdate{AllowContact: b(true)}); err != nil {
		t.Fatalf("update: %v", err)
	}

	// the agency sees a blurred, scored lead and a notification
	leads, err := lsvc.List(ctx, agUser.ID)
	if err != nil || len(leads) != 1 || leads[0].Contact != nil || leads[0].MatchScore < 90 {
		t.Fatalf("leads: %+v %v", leads, err)
	}
	if n, _ := lsvc.Notifications(ctx, agUser.ID); len(n) != 1 {
		t.Fatalf("notifications: %v", n)
	}

	// an unverified agency cannot unlock; a verified one can, once, spending one credit
	if _, err := lsvc.Unlock(ctx, agUser.ID, me.ID, "3.3.3.3"); !errors.Is(err, domain.ErrAgencyNotVerified) {
		t.Fatalf("unverified: %v", err)
	}
	ag, _ := agencies.GetByUser(ctx, agUser.ID)
	if err := pub.VerifyAgency(ctx, hr.ID, ag.ID, true, "4.4.4.4"); err != nil { // hr.ID only serves as an existing user ID for the audit row
		t.Fatalf("verify: %v", err)
	}
	got, err := lsvc.Unlock(ctx, agUser.ID, me.ID, "3.3.3.3")
	if err != nil || got.Contact == nil || got.Contact.Phone != "+57 300 1234567" {
		t.Fatalf("unlock: %+v %v", got, err)
	}
	if _, err := lsvc.Unlock(ctx, agUser.ID, me.ID, "3.3.3.3"); err != nil {
		t.Fatalf("second unlock must be free: %v", err)
	}
	ag, _ = agencies.GetByUser(ctx, agUser.ID)
	if ag.ContactCredits != 4 || !ag.Verified {
		t.Fatalf("credits after one unlock: %+v", ag)
	}
	var audits int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE action = 'agency.unlock_contact' AND record_id = $1`, me.ID).Scan(&audits)
	if audits != 1 {
		t.Fatalf("exactly one unlock audit row expected, got %d", audits)
	}

	// credits can never go negative, even under concurrent unlocks
	if _, err := pool.Exec(ctx, `UPDATE agencies SET contact_credits = 1 WHERE id = $1::text::uuid`, ag.ID); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for k := 0; k < 4; k++ {
		u, err := reg.RegisterNewcomer(ctx, service.Identity{Sub: fmt.Sprintf("auth0|c%d%s", k, run), Email: fmt.Sprintf("c%d%s@x.co", k, run)},
			service.NewcomerRegistration{Consent: consent, FullName: "Car Los", CityOfContract: city, BudgetMax: f(3000000), Bedrooms: i(2), Furnished: b(true)})
		if err != nil {
			t.Fatal(err)
		}
		n, _ := nsvc.Me(ctx, u.ID)
		ids = append(ids, n.ID)
	}
	results := make(chan error, len(ids))
	for _, id := range ids {
		go func(id string) { _, err := lsvc.Unlock(ctx, agUser.ID, id, ""); results <- err }(id)
	}
	okCount := 0
	for range ids {
		if err := <-results; err == nil {
			okCount++
		} else if !errors.Is(err, domain.ErrNoCredits) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	ag, _ = agencies.GetByUser(ctx, agUser.ID)
	if okCount != 1 || ag.ContactCredits != 0 {
		t.Fatalf("one credit must buy exactly one unlock: ok=%d credits=%d", okCount, ag.ContactCredits)
	}

	// favorites and the public listing view
	nu, _ := reg.RegisterNewcomer(ctx, service.Identity{Sub: "auth0|fav" + run, Email: "fav" + run + "@x.co"},
		service.NewcomerRegistration{Consent: service.Consent{Accepted: true}, FullName: "Fav One", CityOfContract: cityPlain})
	if err := psvc.Favorite(ctx, nu.ID, prop.ID); err != nil {
		t.Fatal(err)
	}
	if err := psvc.Favorite(ctx, nu.ID, prop.ID); err != nil {
		t.Fatalf("favoriting twice is idempotent: %v", err)
	}
	if fl, _ := psvc.Favorites(ctx, nu.ID); len(fl) != 1 || fl[0].Address != "" {
		t.Fatalf("favorites: %+v", fl)
	}
	if res, _ := psvc.Search(ctx, cityPlain, nil); len(res) != 1 || res[0].Address != "" {
		t.Fatalf("search must fold accents and hide the address: %+v", res)
	}
	if err := psvc.Delete(ctx, agUser.ID, prop.ID); err != nil {
		t.Fatal(err)
	}

	// public forms
	if err := pub.CreatePartnerInquiry(ctx, service.InquiryInput{FullName: "Carlos", Email: "c@x.co", CollaborationType: "hr", Consent: true}); err != nil {
		t.Fatal(err)
	}
	if err := pub.CreateLead(ctx, service.LeadInput{Name: "Ana", Email: "a@x.co", Consent: true}); err != nil {
		t.Fatal(err)
	}
	if l, err := pub.ListInquiries(ctx); err != nil || len(l) == 0 {
		t.Fatalf("inquiries: %v %v", l, err)
	}
}
