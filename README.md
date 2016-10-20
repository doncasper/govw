Golang Vowpal Wabbit client
===========================

Install
-------

```
$ go get github.com/doncasper/govw
```

Usage
-----

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/doncasper/govw"
)

func main() {
	// First we should create the VW daemon
	vw := govw.NewDaemon("/usr/local/bin/vw", [2]int{26542, 26543}, 10, "/path/to/your.model", true, true)

	// Then we can run VW daemon
	if err := vw.Run(); err != nil {
		log.Fatal("Starting daemon error: ", err)
	}

	// And then we can send data for prediction
	p, err := vw.Predict("1 tag_name| 100:1 200:0.45 250:0.8")
	if err != nil {
		log.Fatal("Predicting error: ", err)
	}

	fmt.Printf("Prediction result: %f | tag: %s\n", p[0].Value, p[0].Tag)
}
```

Stability
---------

At the moment, the client version is not stable and can be changed without notice.
