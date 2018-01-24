# go-autoruns

Collect records of programs registered for persistence on the running system.

## Usage

Invoke the `Autoruns()` function, which will return a slice of `Autorun` structs
with the following properties:

```go
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
```

The values are:

- `Type`: a description of the type of autorun record (e.g. "run_key" or "services").
- `Location`: either a registry key or a file path where the record is stored.
- `ImagePath`: the file path to the executable registered for persistence.
- `ImageName`: just the file name of the executable.
- `Arguments`: any arguments passed to the executable.
- `MD5`: MD5 hash of the executable.
- `SHA1`: SHA1 hash of the executable.
- `SHA256`: SHA256 hash of the executable.

Following is a working example:

```go
package main

import (
	"fmt"
	"github.com/botherder/go-autoruns"
)

func main() {
	autoruns := autoruns.Autoruns()

	for _, autorun := range(autoruns) {
		fmt.Println(autorun.Type)
		fmt.Println(autorun.Location)
		fmt.Println(autorun.ImagePath)
		fmt.Println(autorun.Arguments)
		fmt.Println("")
	}
}
```

## TODO

- Extend support for other autorun records on Windows.
- Extend support for other autorun records on Mac.
- Add support for Linux.
