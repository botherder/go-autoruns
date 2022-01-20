//+build windows
// This file is part of go-autoruns.
// Copyright (c) 2018-2021 Claudio Guarnieri
// See the file 'LICENSE' for copying permission.

package autoruns

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/capnspacehook/taskmaster"
	"golang.org/x/sys/windows/registry"
)

// This function invokes all the platform-dependant functions.
func GetAllAutoruns() (records []*Autorun) {
	records = append(records, WindowsGetCurrentVersionRun()...)
	records = append(records, WindowsGetServices()...)
	records = append(records, WindowsGetStartupFiles()...)
	records = append(records, WindowsGetTasks()...)

	return
}

// This function enumerates items registered through CurrentVersion\Run.
func WindowsGetCurrentVersionRun() (records []*Autorun) {
	regs := []registry.Key{
		registry.LOCAL_MACHINE,
		registry.CURRENT_USER,
	}

	keyNames := []string{
		"Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"Software\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Run",
		"Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
	}

	// We loop through HKLM and HKCU.
	for _, reg := range regs {
		// We loop through the keys we're interested in.
		for _, keyName := range keyNames {
			// Open registry key.
			key, err := registry.OpenKey(reg, keyName, registry.READ)
			if err != nil {
				continue
			}

			// Enumerate value names.
			names, err := key.ReadValueNames(0)
			if err != nil {
				key.Close()
				continue
			}

			for _, name := range names {
				// For each entry we get the string value.
				value, _, err := key.GetStringValue(name)
				if err != nil || value == "" {
					continue
				}

				imageLocation := fmt.Sprintf("%s\\%s", registryToString(reg), keyName)
				// We pass the value string to a function to return an Autorun.
				newAutorun := stringToAutorun("run_key", imageLocation, value, name, true)
				// Add the new autorun to the records.
				records = append(records, newAutorun)
			}
			key.Close()
		}
	}

	return
}

// This function enumerates Windows Services.
func WindowsGetServices() (records []*Autorun) {
	var reg registry.Key = registry.LOCAL_MACHINE
	var servicesKey string = "System\\CurrentControlSet\\Services"

	// Open the registry key.
	key, err := registry.OpenKey(reg, servicesKey, registry.READ)
	if err != nil {
		return
	}

	// Enumerate subkeys.
	names, err := key.ReadSubKeyNames(0)
	key.Close()
	if err != nil {
		return
	}

	for _, name := range names {
		// We open each subkey.
		subkeyPath := fmt.Sprintf("%s\\%s", servicesKey, name)
		subkey, err := registry.OpenKey(reg, subkeyPath, registry.READ)
		if err != nil {
			continue
		}

		// Check if there is an ImagePath value.
		imagePath, _, err := subkey.GetStringValue("ImagePath")
		subkey.Close()
		// If not, we skip to the next one.
		if err != nil {
			continue
		}

		imageLocation := fmt.Sprintf("%s\\%s", registryToString(reg), subkeyPath)
		// We pass the value string to a function to return an Autorun.
		newAutorun := stringToAutorun("service", imageLocation, imagePath, "", true)
		// Add the new autorun to the records.
		records = append(records, newAutorun)
	}

	return
}

// %ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp
// %AppData%\Microsoft\Windows\Start Menu\Programs\Startup
func WindowsGetStartupFiles() (records []*Autorun) {
	// We look for both global and user Startup folders.
	folders := []string{
		os.Getenv("ProgramData"),
		os.Getenv("AppData"),
	}

	// The base path is the same for both.
	var startupBasepath string = "Microsoft\\Windows\\Start Menu\\Programs\\StartUp"

	for _, folder := range folders {
		// Get the full path.
		startupPath := filepath.Join(folder, startupBasepath)

		// Get list of files in folder.
		filesList, err := ioutil.ReadDir(startupPath)
		if err != nil {
			continue
		}

		// Loop through all files in folder.
		for _, fileEntry := range filesList {
			// We skip desktop.ini files.
			if fileEntry.Name() == "desktop.ini" {
				continue
			}

			filePath := filepath.Join(startupPath, fileEntry.Name())
			// Instantiate new autorun record.
			newAutorun := stringToAutorun("startup", startupPath, filePath, "", false)
			// Add new record to list.
			records = append(records, newAutorun)
		}
	}

	return
}

// Extract scheduled tasks that trigger command launches.
func WindowsGetTasks() (records []*Autorun) {
	taskService, err := taskmaster.Connect()
	if err != nil {
		return
	}
	defer taskService.Disconnect()

	tasks, err := taskService.GetRegisteredTasks()
	if err != nil {
		return
	}

	for _, task := range tasks {
		for _, action := range task.Definition.Actions {
			if action.GetType() == taskmaster.TASK_ACTION_EXEC {
				newAutorun := stringToAutorun("task", task.Path, action.(taskmaster.ExecAction).Path, task.Name, true)
				records = append(records, newAutorun)
			}
		}
	}

	return
}
