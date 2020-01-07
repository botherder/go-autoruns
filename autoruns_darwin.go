//+build darwin

package autoruns

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/botherder/go-files"
	"howett.net/plist"
)

type Plist struct {
	Label            string   `plist:"Label"`
	ProgramArguments []string `plist:"ProgramArguments"`
	RunAtLoad        bool     `plist:"RunAtLoad"`
}

func parsePlists(entryType string, folders []string) (records []*Autorun) {
	for _, folder := range folders {
		// Check if the folders exists.
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			continue
		}

		// Get list of files in folder.
		filesList, err := ioutil.ReadDir(folder)
		if err != nil {
			continue
		}

		// Loop through all files in folder.
		for _, fileEntry := range filesList {
			// Open the plist file.
			filePath := filepath.Join(folder, fileEntry.Name())
			reader, err := os.Open(filePath)
			if err != nil {
				continue
			}

			// Parse the plist file.
			var p Plist
			decoder := plist.NewDecoder(reader)
			err = decoder.Decode(&p)
			reader.Close()
			if err != nil {
				continue
			}

			if len(p.ProgramArguments) == 0 {
				continue
			}

			// We skip those that do not start automatically.
			if !p.RunAtLoad {
				continue
			}

			imagePath := p.ProgramArguments[0]
			arguments := ""
			if len(p.ProgramArguments) > 1 {
				arguments = strings.Join(p.ProgramArguments[1:], " ")
			}

			md5, _ := files.HashFile(imagePath, "md5")
			sha1, _ := files.HashFile(imagePath, "sha1")
			sha256, _ := files.HashFile(imagePath, "sha256")

			newAutorun := Autorun{
				Type:         entryType,
				Location:     filePath,
				ImagePath:    imagePath,
				ImageName:    filepath.Base(imagePath),
				Arguments:    arguments,
				MD5:          md5,
				SHA1:         sha1,
				SHA256:       sha256,
				LaunchString: imagePath,
			}
			if arguments != "" {
				newAutorun.LaunchString += " " + arguments
			}

			// Add new record to list.
			records = append(records, &newAutorun)
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
	// Launch when specific user logs in
	launchAgentsUser := []string{}
	if files, err := ioutil.ReadDir("/Users"); err == nil {
		for _, f := range files {
			if f.IsDir() {
				launchAgentsUser = append(launchAgentsUser, filepath.Join("/Users", f.Name(), "Library", "LaunchAgents"))
			}

		}
	}

	records = append(records, parsePlists("launch_daemons", launchDaemons)...)
	records = append(records, parsePlists("launch_agents", launchAgents)...)
	records = append(records, parsePlists("launch_agents_user", launchAgentsUser)...)

	return
}
