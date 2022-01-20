// This file is part of go-autoruns.
// Copyright (c) 2018-2021 Claudio Guarnieri
// See the file 'LICENSE' for copying permission.

package autoruns

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/botherder/go-savetime/hashes"
	"golang.org/x/sys/windows/registry"
)

// Just return a string value for a given registry root Key.
func registryToString(reg registry.Key) string {
	switch reg {
	case registry.LOCAL_MACHINE:
		return "LOCAL_MACHINE"
	case registry.CURRENT_USER:
		return "CURRENT_USER"
	default:
		return ""
	}
}

// cleanPath uses lookPath to search for the correct path to
// the executable and cleans the file path.
func cleanPath(file string) (string, error) {
	file, err := exec.LookPath(file)
	if err != nil {
		return "", err
	}
	return filepath.Clean(file), nil
}

func parsePath(entryValue string) (string, string, error) {
	if entryValue == "" {
		return "", "", errors.New("empty path")
	}
	// do some replacements to convert typical kernel paths to user paths
	if strings.HasPrefix(entryValue, `\??\`) {
		entryValue = entryValue[4:]
	}
	if len(entryValue) >= 11 && strings.ToLower(entryValue[:11]) == "\\systemroot" {
		entryValue = strings.Replace(entryValue, entryValue[:11], os.Getenv("SystemRoot"), -1)
	}
	if len(entryValue) >= 8 && strings.ToLower(entryValue[:8]) == "system32" {
		entryValue = strings.Replace(entryValue, entryValue[:8], fmt.Sprintf("%s\\System32", os.Getenv("SystemRoot")), -1)
	}
	// replace environment variables
	entryValue, err := registry.ExpandString(entryValue)
	if err != nil {
		return "", "", err
	}

	// Now find the executable, analogous to how CreateProcess works
	var executable string
	var arguments string
	if strings.HasPrefix(entryValue, "\"") {
		// Quoted executable - look for closing quote
		closingQuote := strings.Index(entryValue[1:], "\"")
		if closingQuote < 0 {
			return "", "", errors.New("unclosed quote")
		}
		executable = entryValue[1 : closingQuote+1]
		arguments = entryValue[closingQuote+2:]
	} else {
		// Unquoted executable. Try to look for first word first and then extend the path if that fails, e.g.:
		// For C:\Program Files\My Application\app.exe some args, first search for:
		// C:\Program
		// if that fails, look for:
		// C:\Program Files\My
		// And if that still fails, look for:
		// C:\Program Files\My Application\app.exe
		// ...
		var spaceIndex int
		for {
			if spaceIndex == len(entryValue) {
				// Could not find file
				return "", "", errors.New("executable not found")
			}
			if nextSpace := strings.IndexAny(entryValue[spaceIndex+1:], " \t"); nextSpace < 0 {
				spaceIndex = len(entryValue)
			} else {
				spaceIndex += nextSpace + 1
			}
			possibleExecutable := entryValue[:spaceIndex]
			if exePath, err := exec.LookPath(possibleExecutable); err == nil {
				executable = exePath
				if spaceIndex < len(entryValue) {
					arguments = entryValue[spaceIndex+1:]
				}
				break
			}
		}
	}

	arguments = strings.TrimSpace(arguments)
	if v, err := cleanPath(executable); err == nil {
		executable = v
	}
	return executable, arguments, nil
}

func stringToAutorun(entryType string, entryLocation string, entryValue string, entry string, toParse bool) *Autorun {
	var imagePath = entryValue
	var launchString = entryValue
	var argsString = ""

	if toParse {
		executable, args, err := parsePath(entryValue)
		if err == nil {
			imagePath = executable
			argsString = args
		}
	}

	md5, _ := hashes.FileMD5(imagePath)
	sha1, _ := hashes.FileSHA1(imagePath)
	sha256, _ := hashes.FileSHA256(imagePath)

	newAutorun := Autorun{
		Type:         entryType,
		Location:     entryLocation,
		ImagePath:    imagePath,
		ImageName:    filepath.Base(imagePath),
		Arguments:    argsString,
		MD5:          md5,
		SHA1:         sha1,
		SHA256:       sha256,
		Entry:        entry,
		LaunchString: launchString,
	}

	return &newAutorun
}
