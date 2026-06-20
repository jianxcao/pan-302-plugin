package pluginpkg

import "errors"

var (
	ErrInvalidManifest = errors.New("invalid plugin manifest")
	ErrInvalidPackage  = errors.New("invalid plugin package")
)
