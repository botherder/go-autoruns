package autoruns

import (
	"bufio"
	"fmt"
	"github.com/botherder/go-files"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"
)

var interestingBashPatterns = map[string][]string{"$HOME$": {`\..*rc$`}}

/*
Get list of all usernames

:param onlyLoginUser select only users who are able to login
:returns list of usernames
*/
func getUsers(onlyLoginUser bool) (usernames []string) {
	file, err := os.Open("/etc/passwd")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		// skip all line starting with #
		if equal := strings.Index(line, "#"); equal < 0 {
			// get the username and description
			lineSlice := strings.Split(line, ":")

			if len(lineSlice) > 1 {
				if onlyLoginUser && isLoginUser(lineSlice[6]) {
					usernames = append(usernames, lineSlice[0])
				} else {
					usernames = append(usernames, lineSlice[0])
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return usernames
		}
	}

	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	return usernames
}

/*
Is a user with the given shell able to login?

:param shell
:returns is able to login?
*/
func isLoginUser(shell string) (isLoginUser bool) {
	shell = strings.Trim(shell, "\n")
	switch shell {
	case "/usr/bin/nologin", "/bin/nologin", "/bin/false", "/sbin/nologin":
		return false
	default:
		return true
	}
}

/*
Get home directories for the list of usernames

params: list of usernames
returns: home directory for each user
*/
func getHomeDirs(usernames []string) (homeDirectories []string) {
	for _, name := range usernames {

		usr, err := user.Lookup(name)
		if err != nil {
			panic(err)
		}

		homeDirectories = append(homeDirectories, usr.HomeDir)
	}
	return homeDirectories
}

/*
Is the given string a command?

params: name
return: command type
*/
func isCommand(name string) (string, string) {
	// https://stackoverflow.com/a/677212
	resultBytes, _ := exec.Command("/bin/sh", "-c", "command -v "+name).Output()
	result := string(resultBytes)
	result = strings.TrimSuffix(result, "\n")

	if result != "" {
		if strings.Contains(result, "/") {
			return result, "cmd"
		} else {
			return result, "bash"
		}
	} else {
		return "", ""
	}
}

/*
Parse bash file

:params bash file
:returns record of all commands
*/
func parseBashFile(rcFile string) (records []*Autorun) {
	file, err := os.Open(rcFile)

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	autorun := Autorun{
		Location: rcFile,
		Type:     "bash",
	}
	for scanner.Scan() {
		records = append(records, parseBashCommand(scanner.Text(), autorun)...)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	return records
}

/*
Parse Cron File

:params: Location of crond file
:returns record of all commands
*/
func parseCronDFile(cronDFile string) (records []*Autorun) {
	file, err := os.Open(cronDFile)
	usernames := getUsers(false)

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	autorun := Autorun{
		Location: cronDFile,
		Type:     "crond",
	}
	for scanner.Scan() {
		command := getCronCommand(scanner.Text(), usernames)

		parsedCommands := parseBashCommand(command, autorun)
		records = append(records, parsedCommands...)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
	return records
}

/*
Parse bash command

:params text command string
:params baseAutorun Base autorun struct
:returns record of all commands
*/
func parseBashCommand(text string, baseAutorun Autorun) (records []*Autorun) {
	command := strings.TrimSpace(text)
	command += " "
	// filter out comments
	if len(command) == 0 {
		return
	}
	if command[0] == '#' {
		return
	}

	command = removeBashComments(command)
	var applications []string = nil
	newCommands := filterBashCommands(command, false, applications)

	for _, newCommand := range newCommands {
		autorun := baseAutorun
		autorun.LaunchString = command
		autorun.ImagePath = newCommand
		autorun.ImageName = path.Base(autorun.ImagePath)
		autorun.Arguments = strings.SplitN(command, " ", 2)[1]

		autorun.MD5, _ = files.HashFile(autorun.ImagePath, "md5")
		autorun.SHA1, _ = files.HashFile(autorun.ImagePath, "sha1")
		autorun.SHA256, _ = files.HashFile(autorun.ImagePath, "sha256")

		records = append(records, &autorun)
	}

	return records
}

/*
Remove any bash comments

:params command string
:returns cleaned command string
*/
func removeBashComments(command string) (newCommand string) {
	commentRemover, _ := regexp.Compile(`\s#.*$`)
	indices := commentRemover.FindStringIndex(command)
	if len(indices) > 0 {
		return command[:indices[0]]
	} else {
		return command
	}

}

/*
Recursively parse bash command

:params command
:params has the first separator already been found
:params list of application in the bash string
*/
func filterBashCommands(command string, foundFirstSeparator bool, applications []string) []string {
	// Cut off any leading spaces
	command = strings.TrimLeft(command, " ")

	for i, singleRune := range command {
		if foundFirstSeparator == false && singleRune == ' ' {

			commandSplit := strings.Split(command, string(singleRune))
			application, appType := isCommand(commandSplit[0])

			if appType == "cmd" {
				applications = append(applications, application)
			}

			return filterBashCommands(command[i+1:], true, applications)
		} else if isCommandSeparator(singleRune) {
			return filterBashCommands(command[i+1:], false, applications)
		}
	}

	return applications
}

/*
Check if rune is a bash command separator

:params singleRune
:returns is singleRune a bash command separator
*/
func isCommandSeparator(singleRune rune) bool {
	commandSeparator := []string{"|", ";", "&", "\n"}

	for _, separator := range commandSeparator {
		if string(singleRune) == separator {
			return true
		}
	}
	return false
}

/*
Get a list of interesting possible bash files

:returns list of possible bash files to parse
*/
func getBashFiles() (rcFiles []string) {
	placeholders := map[string]func(string) []string{
		"$HOME$": replaceHomePlaceholder,
	}

	for interestingPath, patterns := range interestingBashPatterns {
		for placeholder, replacementFunction := range placeholders {

			improvedPaths := []string{interestingPath}
			if strings.Contains(interestingPath, placeholder) {
				improvedPaths = replacementFunction(interestingPath)
			}

			for _, newPath := range improvedPaths {
				subDirs, _ := ioutil.ReadDir(newPath)

				for _, file := range subDirs {
					filename := file.Name()
					for _, pattern := range patterns {
						matched, _ := regexp.MatchString(pattern, filename)
						if matched {
							filePath := fmt.Sprintf("%s/%s", newPath, filename)
							rcFiles = append(rcFiles, filePath)
						}
					}
				}
			}
		}
	}

	return rcFiles
}

/*
Replace $HOME$ with home directories

:params pattern to replace
:return list of replaced patterns
*/
func replaceHomePlaceholder(pattern string) (patternPaths []string) {
	users := getUsers(true)
	homeDirectories := getHomeDirs(users)

	for _, homeDirectory := range homeDirectories {
		patternPath := strings.Replace(pattern, "$HOME$", path.Dir(homeDirectory+"/"), -1)
		patternPaths = append(patternPaths, patternPath)
	}

	return patternPaths
}

/*
Check if string is present in array
*/
func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

/*
Filter out command from other cron related strings

:params command string and list of usernames
:return cleaned command
*/
func getCronCommand(command string, usernames []string) (newCommand string) {
	command = strings.TrimLeft(command, " ")
	command = strings.ReplaceAll(command, "\t", " ")
	splitCommand := strings.Split(command, " ")
	index := 0

	if len(splitCommand) > 5 {
		if Contains(usernames, splitCommand[5]) {
			index = 6 // because * * * * * root test
		} else {
			index = 5 // because * * * * * test
		}
	} else if hasCronPrefix(splitCommand[0]) {
		index = 1
	}

	return strings.Join(splitCommand[index:], " ")
}

/*
Check if command string starts with a cron prefix

:params command string
:return if command starts with cron prefix
*/
func hasCronPrefix(command string) bool {
	cronPrefixes := []string{
		"@yearly",
		"@annually",
		"@monthly",
		"@weekly",
		"@daily",
		"@hourly",
		"@reboot",
	}

	for _, cronPrefix := range cronPrefixes {
		if strings.HasPrefix(command, cronPrefix) {
			return true
		}
	}
	return false
}
