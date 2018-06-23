package autoruns

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	//"github.com/botherder/go-autoruns"
	files "github.com/botherder/go-files"
)

var section = regexp.MustCompile("\\[.*\\]")

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

func parseUnit(fileName string) (*Autorun, error) {
	fn, err := os.Open(fileName)
	defer fn.Close()

	if err != nil {
		return nil, err
	}

	inSection := ""
	autorun := Autorun{
		Location: fileName,
		Type:     "systemd",
	}
	scanner := bufio.NewScanner(fn)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		// Section
		case section.MatchString(line):
			inSection = line
		case inSection == "[Service]":
			if strings.HasPrefix(line, "ExecStart=") {
				parseShellInvocation(line, &autorun)
			}
		case inSection == "[D-BUS Service]":
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

func findUnits(paths []string) []string {
	units := []string{}
	for x := 0; x < len(paths); x++ {
		path := paths[x]
		filePaths, _ := ioutil.ReadDir(path)
		for y := 0; y < len(filePaths); y++ {
			if strings.HasSuffix(filePaths[y].Name(), ".service") {
				units = append(units, path+filePaths[y].Name())
			}
		}
	}
	return units
}

// GetSystemdAutoruns executes the systemd version of getAutoruns
func GetSystemdAutoruns(paths []string) []*Autorun {
	var records = []*Autorun{}
	units := findUnits(paths)
	for x := 0; x < len(units); x++ {
		unit, err := parseUnit(units[x])
		if err != nil {
			// TODO: Log instead? How should this be handled?
			panic(err)
		}
		records = append(records, unit)
	}
	return records
}
