package runner

import "fmt"

type ErrInvalidTarget string

func (e ErrInvalidTarget) Error() string {
	return fmt.Sprintf("invalid %s", string(e))
}
