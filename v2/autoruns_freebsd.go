// This file is part of go-autoruns.
// Copyright (c) 2018-2021 Claudio Guarnieri
// See the file 'LICENSE' for copying permission.

package autoruns

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
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

	rxp, err := regexp.Compile("^(\\w+)_enable=\"YES\"$")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		matches := rxp.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		serviceName := matches[1]
		if serviceName == "" {
			continue
		}

		enabledServices = append(enabledServices, serviceName)
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
			line = strings.Replace(line, "\"", "", -1)
			matches := rxp.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			serviceName := matches[1]
			if serviceName == "" {
				continue
			}

			for _, enabled := range enabledServices {
				if enabled == serviceName {
					newAutorun := Autorun{
						Type:      entryType,
						Location:  filePath,
						ImageName: serviceName,
					}

					records = append(records, &newAutorun)
					break
				}
			}
		}
	}

	return
}

func GetAllAutoruns() (records []*Autorun) {
	parseRCConf()

	records = append(records, parseRCScripts("rc.d", "/etc/rc.d/")...)
	records = append(records, parseRCScripts("local_rc.d", "/usr/local/etc/rc.d/")...)
	return
}
