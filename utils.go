package govw

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func runCommand(command string, quiet bool) ([]byte, error) {
	switch quiet {
	case false:
		val, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			return []byte{}, err
		}
		return val, nil
	case true:
		err := exec.Command("sh", "-c", command).Start()
		if err != nil {
			return []byte{}, err
		}
		return []byte{}, nil
	default:
		panic("We have some problem with execing command!")
	}
}

// ParsePredictResult get prediction result from VW daemon
// and convert this result into Prediction struct.
func ParsePredictResult(predict *string) *Prediction {
	p := strings.TrimRight(*predict, "\n")

	r := strings.Split(p, " ")

	val, err := strconv.ParseFloat(r[0], 64)
	if err != nil {
		log.Fatal("Error while parsing prediction value:", err)
	}

	if len(r) == 1 {
		return &Prediction{val, ""}
	}

	return &Prediction{val, r[1]}
}

// RecreateDaemon create new VW daemon on another port (default VW port + 1),
// check if all his childrens is wakeup, substitute link to new VW daemon instance.
func RecreateDaemon(d *VWDaemon) {
	log.Println("Start recreating daemon on new port:", d.Port[1])

	port := [2]int{d.Port[1], d.Port[0]}
	newVW := NewDaemon(d.BinPath, port, d.Children, d.Model.Path, d.Test, d.Model.Updatable)
	if err := newVW.Run(); err != nil {
		log.Fatal(err)
	}

	tmpVW := *d
	*d = newVW

	log.Println("Finished recreating daemon on new port:", d.Port[0])
	tmpVW.Stop()
}

// ModelFileChecker check if our model file is changed,
// and recreate VW daemon on a new port.
func ModelFileChecker(vw *VWDaemon) {
	for {
		time.Sleep(time.Second * 1) // TODO: Move count of second to config file

		if vw.Model.IsChanged() {
			log.Println("Model file is changed!")
			RecreateDaemon(vw)
		}
	}
}
