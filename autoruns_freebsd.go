package autoruns

import (
	"bufio"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var enabledServices []string

func parseRCConf() {
	file, err := os.Open("/etc/rc.conf")
	if err != nil {
		return
	}
	defer file.Close()

	rxp, err := regexp.Compile("^(\\w+)_enabled=\"YES\"$")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := rxp.FindString(line)
		if match == "" {
			continue
		}

		enabledServices = append(enabledServices, match)
	}

	if err := scanner.Err(); err != nil {
		return
	}
}

func parseRCScripts(entryType, folder string) (records []*Autorun) {
	// Check if the folders exists.
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return
	}

	// Get list of files in folder.
	filesList, err := ioutil.ReadDir(folder)
	if err != nil {
		return
	}

	rxp, err := regexp.Compile("^name=(\\w+)$")

	// Loop through all files in folder.
	for _, fileEntry := range filesList {
		filePath := filepath.Join(folder, fileEntry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			line = strings.Replace(line, "\"", "")
			name := rxp.FindString(line)
			if name == "" {
				continue
			}

			newAutorun := Autorun{
				Type:      entryType,
				Location:  filePath,
				ImageName: name,
			}

			records = append(records, &newAutorun)
		}
	}

	return
}

func getAutoruns() (records []*Autorun) {
	parseRCConf()

	records = append(records, parseRCScripts("rc.d", "/etc/rc.d/")...)
	records = append(records, parseRCScripts("local_rc.d", "/usr/local/etc/rc.d/")...)
	return
}
