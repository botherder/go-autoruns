package autoruns

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	files "github.com/botherder/go-files"
)

// This function just invokes all the platform-dependant functions.
func getAutoruns() (records []*Autorun) {
	records = append(records, linuxGetSystemd()...)
	records = append(records, linuxGetBashScripts()...)
	records = append(records, getCronFiles()...)
	records = append(records, getCronDFiles()...)

	return
}

var regexSection = regexp.MustCompile("\\[.*\\]")

func parseShellInvocation(shellLine string, autorun *Autorun) {
	autorun.LaunchString = strings.SplitAfter(shellLine, "=")[1]
	// We need to make sure to drop !! from paths
	autorun.ImagePath = strings.Replace(strings.Split(autorun.LaunchString, " ")[0], "!!", "", -1)
	autorun.ImageName = path.Base(autorun.ImagePath)

	args := strings.Split(autorun.LaunchString, " ")
	if len(args) > 1 {
		autorun.Arguments = strings.Join(args[1:], " ")
	}
}

func stringToAutorun(fileName string) (*Autorun, error) {
	reader, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	autorun := Autorun{
		Location: fileName,
		Type:     "systemd",
	}

	inSection := ""
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if regexSection.MatchString(line) {
			inSection = line
		}

		switch inSection {
		case "[Service]":
			if strings.HasPrefix(line, "ExecStart=") {
				parseShellInvocation(line, &autorun)
			}
		case "[D-BUS Service]":
			if strings.HasPrefix(line, "Exec=") {
				parseShellInvocation(line, &autorun)
			}
		}
	}

	autorun.MD5, _ = files.HashFile(autorun.ImagePath, "md5")
	autorun.SHA1, _ = files.HashFile(autorun.ImagePath, "sha1")
	autorun.SHA256, _ = files.HashFile(autorun.ImagePath, "sha256")

	return &autorun, nil
}

func linuxGetSystemd() (records []*Autorun) {
	folders := []string{
		"/etc/systemd/system/",
		"/usr/share/dbus-1/system-services/",
	}

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
			// Skip all files that don't end with .service.
			if !(strings.HasSuffix(fileEntry.Name(), ".service")) {
				continue
			}

			filePath := filepath.Join(folder, fileEntry.Name())
			record, err := stringToAutorun(filePath)
			if err != nil {
				continue
			}

			records = append(records, record)
		}
	}

	return
}

/*
Parse crontab entries for every user and list autorun commands
*/
func getCronFiles() (records []*Autorun) {
	usernames := getUsers(false)
	for _, user := range usernames {
		args := []string{"-u", user, "-l"}
		result, err := exec.Command("/usr/bin/crontab", args...).Output()
		if err == nil {
			for _, line := range strings.Split(string(result), "\n") {
				command := getCronCommand(line, usernames)
				autorun := Autorun{
					Location: "crontab " + user,
					Type:     "cron",
				}
				records = parseBashCommand(command, autorun)
			}
		}
	}
	return records
}

/*
Parse all crond files and list autorun commands
*/
func getCronDFiles() (records []*Autorun) {
	cronFolders := []string{
		"/etc/cron.d",
		"/etc/cron.daily",
		"/etc/cron.hourly",
		"/etc/cron.monthly",
		"/etc/cron.weekly",
	}

	for _, cronFolder := range cronFolders {
		// Get list of files in folder.
		filesList, err := ioutil.ReadDir(cronFolder)
		if err != nil {
			continue
		}

		// Loop through all files in folder.
		for _, fileEntry := range filesList {
			filename := filepath.Join(cronFolder, fileEntry.Name())
			records = append(records, parseCronDFile(filename)...)
		}
	}
	return records
}

/*
Parse all interesting bash files and list autorun commands
*/
func linuxGetBashScripts() (records []*Autorun) {
	bashFiles := getBashFiles()
	for _, bashFile := range bashFiles {
		records = append(records, parseBashFile(bashFile)...)
	}
	return records
}
