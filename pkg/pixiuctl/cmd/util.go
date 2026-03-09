package cmd

func ErrorIsNotFound(err error) bool {
	return err.Error() == "record not found"
}
