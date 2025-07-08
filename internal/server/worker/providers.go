package worker

import (
	"vibe-ddd-golang/internal/application/payment"
	"vibe-ddd-golang/internal/application/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include domain worker modules
	payment.WorkerModule,
	user.WorkerModule,

	// Worker api
	fx.Provide(NewServer),
)
