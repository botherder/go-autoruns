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
	"github.com/mattn/go-shellwords"
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

func parsePath(entryValue string) ([]string, error) {
	if entryValue == "" {
		return nil, errors.New("empty path")
	}
	if strings.HasPrefix(entryValue, `\??\`) {
		entryValue = entryValue[4:]
	}

	// Do some typical replacements.
	if len(entryValue) >= 11 && strings.ToLower(entryValue[:11]) == "\\systemroot" {
		entryValue = strings.Replace(entryValue, entryValue[:11], os.Getenv("SystemRoot"), -1)
	}
	if len(entryValue) >= 8 && strings.ToLower(entryValue[:8]) == "system32" {
		entryValue = strings.Replace(entryValue, entryValue[:8], fmt.Sprintf("%s\\System32", os.Getenv("SystemRoot")), -1)
	}

	// Replace environment variables.
	entryValue, err := registry.ExpandString(entryValue)
	if err != nil {
		return []string{}, err
	}

	// We clean the path for proper backslashes.
	entryValue = strings.Replace(entryValue, "\\", "\\\\", -1)

	// Check if the whole entry is an executable and clean the file path.
	if v, err := cleanPath(entryValue); err == nil {
		return []string{v}, nil
	}

	// Otherwise we can split the entry for executable and arguments
	parser := shellwords.NewParser()
	args, err := parser.Parse(entryValue)
	if err != nil {
		return []string{}, err
	}

	// If the split worked, find the correct path to the executable and clean
	// the file path.
	if len(args) > 0 {
		if v, err := cleanPath(args[0]); err == nil {
			args[0] = v
		}
	}
	return args, nil
}

func stringToAutorun(entryType, entryLocation, entryValue, entry string, toParse bool) *Autorun {
	var imagePath = entryValue
	var launchString = entryValue
	var argsString = ""

	// TODO: This optional parsing is quite spaghetti. To change.
	if toParse {
		args, err := parsePath(entryValue)

		if err == nil && len(args) > 0 {
			imagePath = args[0]
			if len(args) > 1 {
				argsString = strings.Join(args[1:], " ")
			}
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
