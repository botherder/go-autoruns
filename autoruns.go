// This file is part of go-autoruns.
// Copyright (c) 2018-2021 Claudio Guarnieri
// See the file 'LICENSE' for copying permission.

package autoruns

// Autorun contains the details of a program or command found to be launching
// automatically on start-up of the operating system.
type Autorun struct {
	Type         string `json:"type"`
	Location     string `json:"location"`
	ImagePath    string `json:"image_path"`
	ImageName    string `json:"image_name"`
	Arguments    string `json:"arguments"`
	MD5          string `json:"md5"`
	SHA1         string `json:"sha1"`
	SHA256       string `json:"sha256"`
	Entry        string `json:"entry"`
	LaunchString string `json:"launch_string"`
}

// Autoruns returns a list of Autorun items.
func Autoruns() []*Autorun {
	return getAutoruns()
}
