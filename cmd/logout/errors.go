package logout

import "fmt"

const gettingHomeDirErrorTest = "error at getting user home dir: %w"

func newErrorGettingHomeDir(err error) error {
	return fmt.Errorf(gettingHomeDirErrorTest, err)
}
