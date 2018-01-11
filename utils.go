package autoruns

import (
	"strings"
	"path/filepath"
	"github.com/botherder/go-files"
)

func stringToAutorun(entryType string, entryLocation string, entryValue string, toParse bool) *Autorun {
	var imagePath = entryValue
	var argsString = ""

	// TODO: This optional parsing is quite spaghetti. To change.
	if toParse == true {
		args, err := parsePath(entryValue)

		if err == nil {
			imagePath = args[0]
			if len(args) > 1 {
				argsString = strings.Join(args[1:], " ")
			}
		}
	}

	md5, _ := files.HashFile(imagePath, "md5")
	sha1, _ := files.HashFile(imagePath, "sha1")
	sha256, _ := files.HashFile(imagePath, "sha256")

	newAutorun := Autorun{
		Type: entryType,
		Location: entryLocation,
		ImagePath: imagePath,
		ImageName: filepath.Base(imagePath),
		Arguments: argsString,
		MD5: md5,
		SHA1: sha1,
		SHA256: sha256,
	}

	return &newAutorun
}
