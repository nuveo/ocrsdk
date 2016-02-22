ocrsdk
======

ABBYY Cloud OCR SDK

### Install

`go get github.com/poorny/ocrsdk`

### Usage


Export to env `APPLICATION_ID` and `PASSWORD`.

_See [Docs](http://ocrsdk.com/documentation/)_

```go
package main

import (
	"fmt"

	"github.com/poorny/ocrsdk"
)

func main() {
	path := "/path/to/file.pdf"
	lang := "English"

	ocrSDK := ocrsdk.NewProcessImage(os.Getenv("APPLICATION_ID"), os.Getenv("PASSWORD"))
	ocrSDK.Language = "PortugueseBrazilian"
	ocrSDK.Profile = "documentConversion"

	result, err := ocrSDK.Do(fullPath)
	fmt.Println(result)
}
```
