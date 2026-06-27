package database

import (
	"strings"
	"testing"
)

func TestSafeMongoLogTargetRedactsConnectionDetails(t *testing.T) {
	uri := "mongodb://root:strongpassword@mongo.example.test:27017/admin?authSource=admin&tls=true"

	got := safeMongoLogTarget(uri, "statusplatform")

	for _, secret := range []string{
		"root",
		"strongpassword",
		"authSource=admin",
		"tls=true",
		"mongodb://",
		"mongo.example.test",
	} {
		if strings.Contains(got, secret) {
			t.Fatalf("safeMongoLogTarget() = %q, leaked %q", got, secret)
		}
	}

	if !strings.Contains(got, "statusplatform") {
		t.Fatalf("safeMongoLogTarget() = %q, want database name", got)
	}
}
