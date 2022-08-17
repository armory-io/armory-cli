package logout

import "fmt"

const (
	errGettingHomeDirTest = "error at getting user home dir: %w"
)

func newErrorGettingHomeDir(err error) error {
	return fmt.Errorf(errGettingHomeDirTest, err)
}
