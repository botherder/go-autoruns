//+build aix

package autoruns

func parsePath(entryValue string) ([]string, error) {
	return []string{}, nil
}

// This function just invokes all the platform-dependant functions.
func getAutoruns() (records []*Autorun) {
	return
}
