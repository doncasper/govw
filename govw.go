package govw

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// endOfLine is represent byte code for symbol of end of line: `\n`
const endOfLine = 10

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
	Port     [2]int
	Children int
	Model    VWModel
	Test     bool
	TCPQueue chan *net.TCPConn
}

// Predict contain result of prediction
type Prediction struct {
	Value float64
	Tag   string
}

// NewDaemon method return instanse of new Vowpal Wabbit daemon
func NewDaemon(binPath string, ports [2]int, children int, modelPath string, test bool, updatable bool) VWDaemon {
	info, err := os.Stat(modelPath)
	if err != nil {
		log.Fatal(err)
	}

	return VWDaemon{
		BinPath:  binPath,
		Port:     ports,
		Children: children,
		Model:    VWModel{modelPath, info.ModTime(), updatable},
		Test:     test,
	}
}

func (vw *VWDaemon) getTCPConn() *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", vw.Port[0]))
	if err != nil {
		log.Fatal("Error while resolving IP addr: ", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal("Error while dialing TCP", err)
	}

	return conn
}

func (vw *VWDaemon) makeTCPConnQueue() {
	log.Println("Start creating TCP connections queue:", len(vw.TCPQueue))
	size := vw.Children / 2

	for i := 0; i < size; i++ {
		vw.TCPQueue <- vw.getTCPConn()
	}

	log.Println("Queue of TCP connections for VW is created:", len(vw.TCPQueue))
}

// Run method send command for starting new VW daemon.
func (vw *VWDaemon) Run() error {
	if vw.IsExist(3, 100) {
		vw.Stop()
	}

	cmd := fmt.Sprintf("vw --daemon --threads --quiet --port %d --num_children %d", vw.Port[0], vw.Children)

	if vw.Model.Path != "" {
		cmd += fmt.Sprintf(" -i  %s", vw.Model.Path)
	}

	if vw.Test {
		cmd += " -t"
	}

	if _, err := runCommand(cmd, true); err != nil {
		panic(err)
	}

	if !vw.IsExist(5, 500) {
		log.Fatal("Failed to start daemon!")
	}

	log.Printf("Vowpal wabbit daemon is running on port: %d", vw.Port[0])

	vw.TCPQueue = make(chan *net.TCPConn, vw.Children/2)
	vw.makeTCPConnQueue()

	return nil
}

// Stop current daemon
func (vw *VWDaemon) Stop() error {
	log.Println("Try stop daemon on port:", vw.Port[0])

	cmd := fmt.Sprintf("pkill -9 -f \"vw.*--port %d\"", vw.Port[0])
	if _, err := runCommand(cmd, true); err != nil {
		panic(err)
	}

	if vw.IsExist(10, 500) {
		log.Fatal("Failed to stop daemon!")
	}

	log.Println("Stoped VW daemon on port:", vw.Port[0])

	log.Printf("Start closing TCP connections to: %d (%d)\n", vw.Port[0], len(vw.TCPQueue))
	for i := 0; i < len(vw.TCPQueue); i++ {
		conn := <-vw.TCPQueue
		conn.Close()
	}
	log.Println("End closing TCP connections to:", vw.Port[0])

	return nil
}

// Predict method get predictions strings then send data to VW daemon
// for getting prediction result and return list of predictions result.
func (vw *VWDaemon) Predict(pData ...string) ([]*Prediction, error) {
	size := len(pData)
	result := make([]*Prediction, size)

	data := []byte(strings.Join(pData, "\n"))
	data = append(data, endOfLine)

	failed := false

	for {
		conn := <-vw.TCPQueue

		if _, err := conn.Write(data); err != nil {
			log.Println("Error while writing to VW TCP connections: ", err, vw.Port[0])
			continue
		}

		reader := bufio.NewReader(conn)

		for i := 0; i < size; i++ {
			res, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() != "EOF" {
					log.Println("Error while reading VW response: ", err, vw.Port[0], conn.RemoteAddr())
					log.Println("Retry predict value!")
				}

				failed = true
				break
			}

			result[i] = ParsePredictResult(&res)
		}

		if failed {
			failed = false
			continue
		}

		if strings.HasSuffix(conn.RemoteAddr().String(), strconv.Itoa(vw.Port[0])) {
			vw.TCPQueue <- conn
		} else {
			conn.Close()
		}

		return result, nil
	}
}

func (vw *VWDaemon) WorkersCount() (int, error) {
	cmd := fmt.Sprintf("pgrep -f 'vw.*--port %d' | wc -l", vw.Port[0])
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

	//log.Println("Start checking IsExist!")
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

// IsChanged method checks whether the model file has been modified.
func (model *VWModel) IsChanged() (bool, error) {
	info, err := os.Stat(model.Path)
	if err != nil {
		log.Println(err)
		return false, err
	}

	if model.ModTime != info.ModTime() {
		return true, nil
	}

	return false, nil
}
