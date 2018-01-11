package autoruns

import (
	"os"
	"fmt"
	"regexp"
	"strings"
	"io/ioutil"
	"path/filepath"
	"golang.org/x/sys/windows/registry"
	"github.com/mattn/go-shellwords"
)

// Just return a string value for a given registry root Key.
func registryToString(reg registry.Key) string {
	if reg == registry.LOCAL_MACHINE {
		return "LOCAL_MACHINE"
	} else if reg == registry.CURRENT_USER {
		return "CURRENT_USER"
	} else {
		return ""
	}
}

func parsePath(entryValue string) ([]string, error) {
	// We clean the path to introduce environment variables.
	re := regexp.MustCompile(`\%(?i)SystemRoot\%`)
	entryValue = re.ReplaceAllString(entryValue, os.Getenv("SystemRoot"))
	re = regexp.MustCompile(`\\(?i)SystemRoot`)
	entryValue = re.ReplaceAllString(entryValue, os.Getenv("SystemRoot"))
	re = regexp.MustCompile(`\%(?i)ProgramFiles\%`)
	entryValue = re.ReplaceAllString(entryValue, os.Getenv("ProgramFiles"))

	// We clean the path for proper backslashes.
	entryValue = strings.Replace(entryValue, "\\", "\\\\", -1)

	// Parse the value to separate path with arguments.
	// TODO: this is super spaghetti and doesn't actually always work. Fix please.
	parser := shellwords.NewParser()
	parser.ParseEnv = true
	args, err := parser.Parse(entryValue)
	if err != nil {
		return []string{}, err
	}

	return args, nil
}

// This function invokes all the platform-dependant functions.
func getAutoruns() (records []*Autorun) {
	records = append(records, windowsGetCurrentVersionRun()...)
	records = append(records, windowsGetServices()...)
	records = append(records, windowsGetStartupFiles()...)
	// records = append(records, windowsGetTasks()...)

	return
}

// This function enumerates items registered through CurrentVersion\Run.
func windowsGetCurrentVersionRun() (records []*Autorun) {
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
	for _, reg := range(regs) {
		// We loop through the keys we're interested in.
		for _, keyName := range(keyNames) {
			// Open registry key.
			key, err := registry.OpenKey(reg, keyName, registry.READ)
			if err != nil {
				continue
			}

			// Enumerate value names.
			names, err := key.ReadValueNames(0)
			if err != nil {
				continue
			}

			for _, name := range names {
				// For each entry we get the string value.
				value, _, err := key.GetStringValue(name)
				if err != nil {
					continue
				}

				imageLocation := fmt.Sprintf("%s\\%s", registryToString(reg), keyName)

				// We pass the value string to a function to return an Autorun.
				newAutorun := stringToAutorun("run_key", imageLocation,  value, true)

				// Add the new autorun to the records.
				records = append(records, newAutorun,)
			}
		}
	}

	return
}

// This function enumerates Windows Services.
func windowsGetServices() (records []*Autorun) {
	var reg registry.Key = registry.LOCAL_MACHINE
	var servicesKey string = "System\\CurrentControlSet\\Services"

	// Open the registry key.
	key, err := registry.OpenKey(reg, servicesKey, registry.READ)
	if err != nil {
		return
	}

	// Enumerate subkeys.
	names, err := key.ReadSubKeyNames(0)
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
		// If not, we skip to the next one.
		if err != nil {
			continue
		}

		imageLocation := fmt.Sprintf("%s\\%s", registryToString(reg), subkeyPath)

		// We pass the value string to a function to return an Autorun.
		newAutorun := stringToAutorun("service", imageLocation, imagePath, true)

		// Add the new autorun to the records.
		records = append(records, newAutorun,)
	}

	return
}

// %ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp
// %AppData%\Microsoft\Windows\Start Menu\Programs\Startup
func windowsGetStartupFiles() (records []*Autorun) {
	// We look for both global and user Startup folders.
	folders := []string{
		os.Getenv("ProgramData"),
		os.Getenv("AppData"),
	}

	// The base path is the same for both.
	var startupBasepath string = "Microsoft\\Windows\\Start Menu\\Programs\\StartUp"

	for _, folder := range(folders) {
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
			newAutorun := stringToAutorun("startup", startupPath, filePath, false)

			// Add new record to list.
			records = append(records, newAutorun,)
		}
	}

	return
}

// func windowsGetTasks() {

// }
