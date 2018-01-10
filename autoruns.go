package autoruns

type Autorun struct {
	Type		string `json:"type"`
	Location	string `json:"location"`
	ImagePath	string `json:"image_path"`
	ImageName	string `json:"image_name"`
	Arguments	string `json:"arguments"`
	MD5 		string `json:"md5"`
	SHA1		string `json:"sha1"`
	SHA256		string `json:"sha256"`
}

func Autoruns() []*Autorun {
	return getAutoruns()
}
