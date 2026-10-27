package domain

import "testing"

func TestValidEmail(t *testing.T) {
	ok := []string{"a@b.co", "ana.silva+x@empresa.com.br"}
	bad := []string{"", "a@b", "Name <a@b.co>", "a b@c.com", "a@b.co,c@d.com", "@b.co", "a@@b.co"}
	for _, e := range ok {
		if !ValidEmail(e) {
			t.Errorf("%q should be valid", e)
		}
	}
	for _, e := range bad {
		if ValidEmail(e) {
			t.Errorf("%q should be invalid", e)
		}
	}
}

func TestValidPhoneDayCurrencyURL(t *testing.T) {
	if !ValidPhone("+57 300 123-4567") || ValidPhone("abc") || ValidPhone("123") || ValidPhone("12345678901234567890123") {
		t.Fatal("phone")
	}
	if !ValidDay("2026-02-28") || ValidDay("2026-02-30") || ValidDay("28/02/2026") {
		t.Fatal("day")
	}
	if !ValidCurrency("COP") || ValidCurrency("cop") || ValidCurrency("COPS") {
		t.Fatal("currency")
	}
	if !ValidHTTPURL("https://x.co/a") || ValidHTTPURL("javascript:alert(1)") || ValidHTTPURL("ftp://x.co") {
		t.Fatal("url")
	}
}

func TestCleanText(t *testing.T) {
	if got := CleanText("  a\nb\x00c  ", 10); got != "a bc" {
		t.Fatalf("got %q", got)
	}
	if got := CleanText("ááááá", 3); got != "ááá" {
		t.Fatalf("rune cut: %q", got)
	}
	l := CleanList([]string{" Chapinero", "chapinero", "", "Usaquén"}, 10, 50)
	if len(l) != 2 {
		t.Fatalf("dedupe: %v", l)
	}
}

func TestRoles(t *testing.T) {
	if RoleAdmin.SelfRegistrable() || !RoleAgency.SelfRegistrable() || UserRole("root").Valid() {
		t.Fatal("roles")
	}
}
