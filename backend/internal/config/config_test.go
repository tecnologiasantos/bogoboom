package config

import "testing"

func TestProductionRefusesUnsafeSettings(t *testing.T) {
	ok := func() *Config {
		return &Config{Env: "production", Server: ServerConfig{FrontendURL: "https://bogoboom.com"},
			Database: &DatabaseConfig{SSLMode: "require", Password: "x"}, Auth0: Auth0Config{Domain: "t.auth0.com", Audience: "api"}}
	}
	if err := ok().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*Config){
		"http frontend": func(c *Config) { c.Server.FrontendURL = "http://bogoboom.com" },
		"ssl disabled":  func(c *Config) { c.Database.SSLMode = "disable" },
		"no password":   func(c *Config) { c.Database.Password = "" },
		"no auth0":      func(c *Config) { c.Auth0.Domain = " " },
	} {
		c := ok()
		mut(c)
		if c.Validate() == nil {
			t.Errorf("%s must be refused", name)
		}
	}
	dev := &Config{Env: "development", Server: ServerConfig{FrontendURL: "http://localhost:4200"},
		Database: &DatabaseConfig{SSLMode: "disable"}, Auth0: Auth0Config{Domain: "t.auth0.com", Audience: "api"}}
	if err := dev.Validate(); err != nil {
		t.Fatalf("development may use plain http and no ssl: %v", err)
	}
}

func TestConnectionStringEncodesSecrets(t *testing.T) {
	c := &DatabaseConfig{Host: "db", Port: "5432", User: "u@x", Password: "p/ss:w#rd", DBName: "bogoboom", SSLMode: "require"}
	got := c.ConnectionString()
	want := "postgres://u%40x:p%2Fss%3Aw%23rd@db:5432/bogoboom?sslmode=require"
	if got != want {
		t.Fatalf("got %s", got)
	}
}
