package database

import (
	"strings"
	"testing"

	"vibe-ddd-golang/internal/common/enum"
)

func TestGuardAutoMigrateBlocksProductionLikeEnvs(t *testing.T) {
	cases := []struct {
		name  string
		env   enum.EnvEnum
		block bool
	}{
		{"production", enum.PRODUCTION, true},
		{"staging", enum.STAGING, true},
		{"development allowed", enum.DEVELOPMENT, false},
		{"local allowed", enum.LOCAL, false},
		{"empty allowed", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := guardAutoMigrate(tc.env)
			if tc.block {
				if err == nil {
					t.Fatalf("expected guardAutoMigrate to reject environment=%q, got nil", tc.env)
				}
				if !strings.Contains(err.Error(), "AutoMigrate forbidden") {
					t.Fatalf("expected error to mention 'AutoMigrate forbidden', got: %v", err)
				}
				if !strings.Contains(err.Error(), "make migrate-diff") {
					t.Fatalf("expected error to point at Atlas (make migrate-diff), got: %v", err)
				}
			} else if err != nil {
				t.Fatalf("expected guardAutoMigrate to allow environment=%q, got: %v", tc.env, err)
			}
		})
	}
}
