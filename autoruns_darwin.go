package autoruns

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	"howett.net/plist"
)

type Plist struct {
	Label string `plist:"Label"`
	ProgramArguments []string `plist:"ProgramArguments"`
}

func parsePath(entryValue string) ([]string, error) {
	return []string{}, nil
}

func parsePlists(recordType string, folders []string) (records []*Autorun) {
	for _, folder := range folders {
		// Check if the folders exists.
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			continue
		}

		// Get list of files in folder.
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			continue
		}

		// Loop through all files in folder.
		for _, file := range files {
			// Open the plist file.
			filePath := filepath.Join(folder, file.Name())
			reader, err := os.Open(filePath)
			if err != nil {
				continue
			}
			defer reader.Close()

			// Parse the plist file.
			var p Plist
			decoder := plist.NewDecoder(reader)
			err = decoder.Decode(&p)
			if err != nil {
				continue
			}

			if len(p.ProgramArguments) == 0 {
				continue
			}

			// TODO: this is some spaghetti to generate the Autorun record.
			// To change.
			entryValue := strings.Join(p.ProgramArguments[:], " "))
			newAutorun := stringToAutorun(recordType, folder, entryValue, true)

			// Add new record to list.
			records = append(records, newAutorun,)
		}
	}

	return
}

// This function just invokes all the platform-dependant functions.
func getAutoruns() (records []*Autorun) {
	// Startup and run as root.
	launchDaemons := []string{
		"/Library/LaunchDaemons",
		"/System/Library/LaunchDaemons",
	}
	// Launch when any user logs in.
	launchAgents := []string{
		"/Library/LaunchAgents",
		"/System/Library/LaunchAgents",
	}
	// Launch when current user logs in.
	launchAgentsUser := []string{
		filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents"),
	}

	records = append(records, parsePlists("launch_daemons", launchDaemons)...)
	records = append(records, parsePlists("launch_agents", launchAgents)...)
	records = append(records, parsePlists("launch_agents_user", launchAgentsUser)...)

	return
}
