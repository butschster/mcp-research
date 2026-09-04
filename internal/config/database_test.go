package config

import "testing"

func TestDatabaseInMemory(t *testing.T) {
	for _, tc := range []struct {
		cfg  Config
		want bool
	}{
		{Config{}, true}, {Config{DBPath: ":memory:"}, true},
		{Config{DBDSN: "file:shared?mode=memory&cache=shared"}, true},
		{Config{DBPath: "research.db"}, false},
		{Config{DBPath: "old.db", DBDSN: ":memory:"}, true},
		{Config{DBDSN: "research.db"}, false},
		{Config{DBDSN: "research-mode=memory.db"}, false},
		{Config{DBDriver: "postgres", DBDSN: "postgres://localhost/db"}, false},
		{Config{DBDriver: "mysql"}, false},
	} {
		if got := tc.cfg.DatabaseInMemory(); got != tc.want {
			t.Errorf("%+v: got %v want %v", tc.cfg, got, tc.want)
		}
	}
}
