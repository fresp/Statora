package database

import "testing"

func TestDriverIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		got  Driver
		want string
	}{
		{name: "mongodb", got: DriverMongoDB, want: "mongodb"},
		{name: "postgres", got: DriverPostgres, want: "postgres"},
		{name: "mysql", got: DriverMySQL, want: "mysql"},
		{name: "sqlite", got: DriverSQLite, want: "sqlite"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("driver identifier = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestIsSupportedDriver(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		want   bool
	}{
		{name: "mongodb", driver: "mongodb", want: true},
		{name: "postgres", driver: "postgres", want: true},
		{name: "mysql", driver: "mysql", want: true},
		{name: "sqlite", driver: "sqlite", want: true},
		{name: "unknown", driver: "oracle", want: false},
		{name: "empty", driver: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedDriver(tt.driver); got != tt.want {
				t.Fatalf("IsSupportedDriver(%q) = %t, want %t", tt.driver, got, tt.want)
			}
		})
	}
}
