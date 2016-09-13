Golang Vowpal Wabbit client
===========================

Usage
-----

```golang
package main

import (
	"fmt"
	"log"
	
	"github.com/DonCasper/govw"
)

func main() {
	// First we should create the VW daemon
	vw := govw.NewDaemon("/usr/local/bin/vw", 26542, 10, "/path/to/your.model", true, true, true)

	// Then we can run VW daemon
	if err := vw.Run(); err != nil {
		log.Fatal("Starting daemon error: ", err)
	}

	// And then we can send data for prediction
	p, err := vw.Predict([]byte("0 | 100:1 200:0.45 250:0.8"))
	if err != nil {
		log.Fatal("Predicting error: ", err)
	}

	fmt.Println("Prediction result:", p)
}
```

Stability
---------

At the moment, the client version is not stable and can be changed without notice.
