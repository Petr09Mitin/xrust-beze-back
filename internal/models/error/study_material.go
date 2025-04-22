package custom_errors

import (
	"fmt"
)

var (
	ErrNoMaterialAuthorID = fmt.Errorf("%w: no material author_id provided", ErrBadRequest)
	ErrNoMaterialID       = fmt.Errorf("%w: no material id provided", ErrBadRequest)
	ErrNoMaterialTag      = fmt.Errorf("%w: no material tag provided", ErrBadRequest)
	ErrNoMaterialName     = fmt.Errorf("%w: no material name provided", ErrBadRequest)
)
