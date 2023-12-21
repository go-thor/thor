// Package build contains application build information
// ```sh
// go run -ldflags="-X 'github.com/go-thor/thor/build.Namespace={Namespace}' -X 'github.com/go-thor/thor/build.Name={Name}' -X 'github.com/go-thor/thor/build.Version={Version}' -X 'github.com/go-thor/thor/build.Instance={Instance}' -X 'github.com/go-thor/thor/build.BuildId={BuildId}' -X 'github.com/go-thor/thor/build.BuildTime={BuildTime}'" .
// ```
package build

import (
	"os"
	"strings"
)

var (
	Namespace = ""
	Name      = ""
	Version   = ""
	Instance  = ""
	BuildId   = ""
	BuildTime = ""
)

func init() {
	if Instance == "" {
		Instance, _ = os.Hostname()
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

func Info() string {
	return strings.Join([]string{
		"Namespace: " + Namespace,
		"Name: " + Name,
		"Version: " + Version,
		"Instance: " + Instance,
		"BuildId: " + BuildId,
		"BuildTime: " + BuildTime,
	}, "\n")
}
