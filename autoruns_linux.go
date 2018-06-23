package autoruns

func parsePath(entryValue string) ([]string, error) {
	return []string{}, nil
}

// This function just invokes all the platform-dependant functions.
func getAutoruns() (records []*Autorun) {
	systemdRecords := GetSystemdAutoruns([]string{"/etc/systemd/system/", "/usr/share/dbus-1/system-services/"})
	records = append(records, systemdRecords...)
	return records
}
