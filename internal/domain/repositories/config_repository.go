package repositories

import "github.com/connect0459/edit-pr-duration/internal/domain/entities"

// ConfigRepository abstracts loading of configuration
type ConfigRepository interface {
	// Load reads configuration from the given path
	//
	// Arguments:
	//   - path: path to the config file
	//
	// Returns:
	//   - the config object
	//   - an error, if any
	Load(path string) (*entities.Config, error)
}
