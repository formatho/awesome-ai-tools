package agentrunner

import "fmt"

// ErrAgentAlreadyExists is returned when trying to create an agent that already exists.
var ErrAgentAlreadyExists = fmt.Errorf("agent already exists")

// ErrInvalidConfig is returned when configuration validation fails.
type ErrInvalidConfig map[string]string

func (e ErrInvalidConfig) Error() string {
	msg := "invalid configuration:"
	for field, err := range e {
		msg += fmt.Sprintf(" %s=%s", field, err)
	}
	return msg
}

// IsErrInvalidConfig checks if an error is of type ErrInvalidConfig.
func IsErrInvalidConfig(err error) bool {
	_, ok := err.(ErrInvalidConfig)
	return ok
}
