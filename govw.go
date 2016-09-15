package govw

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultPort = 26542

// VWModel contain information about VW model file
// If `Updatable` field is `true`, the system will be track of the
// changes model file and restart the daemon if necessary
type VWModel struct {
	Path      string
	ModTime   time.Time
	Updatable bool
}

// VWDaemon contain information about VW daemon
type VWDaemon struct {
	BinPath  string
	Port     int
	Children int
	Model    *VWModel
	Test     bool
	Quite    bool
	TCPConn  *net.TCPConn
}

// Predict contain result of prediction
type Predict struct {
	Value float64
	Tag   string
}

func NewDaemon(binPath string, port int, children int, modelPath string, test bool, quite bool, updatable bool) *VWDaemon {
	info, err := os.Stat(modelPath)
	if err != nil {
		log.Fatal(err)
	}

	daemon := &VWDaemon{
		BinPath:  binPath,
		Port:     port,
		Children: children,
		Model:    &VWModel{modelPath, info.ModTime(), updatable},
		Test:     test,
		Quite:    quite,
	}

	if updatable {
		go modelFileChecker(daemon)
	}

	return daemon
}

func (vw *VWDaemon) getTCPConn() *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", vw.Port))
	if err != nil {
		log.Fatal("Error via resolve IP addr: ", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal("Error via dial TCP", err)
	}

	return conn
}

// Run method send command for starting new VW daemon.
func (vw *VWDaemon) Run() error {
	if vw.IsExist(3, 100) {
		vw.Stop()
	}

	cmd := fmt.Sprintf("vw --daemon --port %d --num_children %d", vw.Port, vw.Children)

	if vw.Model.Path != "" {
		cmd += fmt.Sprintf(" -i  %s", vw.Model.Path)
	}

	if vw.Test {
		cmd += " -t"
	}

	if vw.Quite {
		cmd += " --quiet"
	}

	if _, err := runCommand(cmd, true); err != nil {
		panic(err)
	}

	if !vw.IsExist(5, 500) {
		log.Fatal("Failed to start daemon!")
	}

	vw.TCPConn = vw.getTCPConn()

	log.Printf("Vowpal wabbit daemon is running on port: %d", vw.Port)

	return nil
}

// Stop current daemon
func (vw *VWDaemon) Stop() error {
	cmd := fmt.Sprintf("pkill -9 -f \"vw.*--port %d\"", vw.Port)
	if _, err := runCommand(cmd, true); err != nil {
		panic(err)
	}

	if vw.IsExist(5, 500) {
		log.Fatal("Failed to stop daemon!")
	}

	return nil
}

// Predict method get slice of bytes (you should convert your predict string to bytes),
// then send data to VW daemon for getting prediction result.
func (vw *VWDaemon) Predict(pData []byte) (*Predict, error) {
	// Check if we have `\n` symbol in the end of prediction string
	if pData[len(pData)-1] != 10 {
		pData = append(pData, 10)
	}

	_, err := vw.TCPConn.Write(pData)
	if err != nil {
		log.Fatal("Error via writing to VW TCP connections: ", err)
	}

	res, err := bufio.NewReader(vw.TCPConn).ReadString('\n')
	if err != nil {
		log.Fatal("Error via reading VW response: ", err)
	}

	return parsePredictResult(&res), nil
}

func (vw *VWDaemon) WorkersCount() (int, error) {
	cmd := fmt.Sprintf("pgrep -f 'vw.*--port %d' | wc -l", vw.Port)
	res, err := runCommand(cmd, false)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(strings.Trim(string(res), "\n"))
	if err != nil {
		return 0, err
	}

	// We should substract 1 from count, to get clear result without
	// side effect of using `sh -c` command in `exec.Command`.
	return count - 1, nil
}

// IsExist method checks if VW daemon and all of his childrens is running.
// You shoud defain count of tries and delay in milliseconds between each try.
func (vw *VWDaemon) IsExist(tries int, delay int) bool {
	var count int
	var err error

	log.Println("Start checking IsExist!")
	for i := 0; i < tries; i++ {
		count, err = vw.WorkersCount()

		// We add 1 to `vw.children`, because we still have the parent process.
		if count == vw.Children+1 {
			return true
		}

		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	if err != nil {
		log.Fatal("Can't getting VW workers count.", err)
	}

	return false
}

func (vw *VWDaemon) DeepCopy() *VWDaemon {
	var copyBuffer bytes.Buffer
	var newVW VWDaemon

	enc := gob.NewEncoder(&copyBuffer)
	err := enc.Encode(vw)
	if err != nil {
		log.Fatal("Deep copy encode:", err)
	}

	dec := gob.NewDecoder(&copyBuffer)
	err = dec.Decode(&newVW)
	if err != nil {
		log.Fatal("Deep copy decode:", err)
	}

	return &newVW
}

// IsChanged method checks whether the model file has been modified.
func (model *VWModel) IsChanged() bool {
	info, err := os.Stat(model.Path)
	if err != nil {
		panic(err)
	}

	if model.ModTime != info.ModTime() {
		return true
	}

	return false
}
