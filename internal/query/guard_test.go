package query

import "testing"

func TestInspect(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		action       string
		needsConfirm bool
	}{
		{name: "select passes", sql: "SELECT * FROM users", action: "SELECT", needsConfirm: false},
		{name: "insert passes", sql: "INSERT INTO users (a) VALUES (1)", action: "INSERT", needsConfirm: false},
		{name: "update with where passes", sql: "UPDATE users SET a = 1 WHERE id = 2", action: "UPDATE", needsConfirm: false},
		{name: "delete without where needs confirm", sql: "DELETE FROM users", action: "DELETE", needsConfirm: true},
		{name: "update without where needs confirm", sql: "UPDATE users SET a = 1", action: "UPDATE", needsConfirm: true},
		{name: "drop needs confirm", sql: "DROP TABLE users", action: "DROP", needsConfirm: true},
		{name: "truncate needs confirm", sql: "TRUNCATE users", action: "TRUNCATE", needsConfirm: true},
		{name: "comments stripped before classify", sql: "-- remove all\nDELETE FROM users", action: "DELETE", needsConfirm: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := Inspect(tt.sql)
			if err != nil {
				t.Fatalf("Inspect(%q) error = %v", tt.sql, err)
			}
			if decision.Action != tt.action {
				t.Errorf("Inspect(%q).Action = %q, want %q", tt.sql, decision.Action, tt.action)
			}
			if decision.NeedsConfirm != tt.needsConfirm {
				t.Errorf("Inspect(%q).NeedsConfirm = %v, want %v", tt.sql, decision.NeedsConfirm, tt.needsConfirm)
			}
		})
	}
}
