// Package build contains application build information
// ```sh
// go build -ldflags="-X 'github.com/go-thor/thor/build.ID={ID}' -X 'github.com/go-thor/thor/build.Name={Name}' -X 'github.com/go-thor/thor/build.Version={Version}' -X 'github.com/go-thor/thor/build.Namespace={Namespace}'"
// ```
package build

import (
	"os"
)

var (
	Namespace = ""
	Name      = ""
	Version   = ""
	ID        = ""
)

func init() {
	if ID == "" {
		ID, _ = os.Hostname()
	}

	if Name == "" {
		Name = "default"
	}

	if Namespace == "" {
		Namespace = "default"
	}

	if Version == "" {
		Version = "0.0.0"
	}
}
