package api

import (
	"vibe-ddd-golang/internal/application/payment"
	"vibe-ddd-golang/internal/application/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include all domain modules
	user.Module,
	payment.Module,

	// API api
	fx.Provide(NewServer),
)
