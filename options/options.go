package options

import (
	"context"
)

type Options struct {
	ServiceName         string
	Address             string
	CertContent         string
	AuthenticationToken string

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}
