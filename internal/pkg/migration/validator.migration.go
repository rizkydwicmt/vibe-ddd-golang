package migration

import (
	"fmt"
)

type FlagOptions struct {
	Init       bool
	Diff       bool
	DryRun     bool
	Name       string
	Force      bool
	Verbose    bool
	Apply      bool
	Baseline   bool
	ApplyCount int
	Rollback   bool
	Version    string
	DevDSN     string
	Status     bool
	Help       bool
}

func ValidateFlags(opts *FlagOptions) error {
	if opts.Rollback {
		if opts.Init || opts.Diff || opts.Apply || opts.Baseline {
			return fmt.Errorf("--rollback cannot be used with --init, --diff, --apply, or --baseline flags")
		}
		return nil
	}

	if opts.Status {
		if opts.Init || opts.Diff || opts.Apply || opts.Rollback || opts.Baseline {
			return fmt.Errorf("--status cannot be used with --init, --diff, --apply, --rollback, or --baseline flags")
		}
		return nil
	}

	if opts.Baseline {
		if opts.Init || opts.Diff || opts.Rollback {
			return fmt.Errorf("--baseline cannot be used with --init, --diff, or --rollback flags")
		}
		return nil
	}

	if !opts.Init && !opts.Diff && !opts.Rollback && !opts.Apply && !opts.Status && !opts.Baseline {
		return fmt.Errorf("please specify either --init, --diff, --rollback, --apply, --baseline, or --status flag")
	}

	if opts.Force && !opts.Apply && !opts.Init && !opts.Baseline {
		return fmt.Errorf("--force can only be used with --apply, --init, or --baseline")
	}
	if opts.Init && opts.Diff {
		return fmt.Errorf("cannot specify both --init and --diff flags")
	}
	if opts.Name != "" && !IsValidMigrationName(opts.Name) {
		return fmt.Errorf("invalid migration name: must contain only letters, numbers, and underscores")
	}
	if opts.ApplyCount < 0 {
		return fmt.Errorf("--count must be a non-negative number")
	}
	if opts.ApplyCount > 0 && !opts.Apply {
		return fmt.Errorf("--count can only be used with --apply")
	}
	if opts.Version != "" && !opts.Rollback && !opts.Baseline {
		return fmt.Errorf("--version can only be used with --rollback or --baseline")
	}
	return nil
}

func IsValidMigrationName(name string) bool {
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}
