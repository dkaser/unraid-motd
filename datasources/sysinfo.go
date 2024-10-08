package datasources

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cosandr/go-motd/utils"
)

type ConfSysInfo struct {
	ConfBase `yaml:",inline"`
}

func (c *ConfSysInfo) Init() {
	c.ConfBase.Init()
	c.PadHeader = []int{0, 3}
	c.PadContent = []int{0, 0}
}

// GetSysInfo various stats about the host Linux OS (kernel, distro, load and more)
func GetSysInfo(ch chan<- SourceReturn, conf *Conf) {
	c := conf.SysInfo
	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	type entry struct {
		name    string
		content string
	}
	// Fetch all the things
	var info = [...]entry{
		{"Version", getDistroName()},
		{"Kernel", getKernel()},
		{"Uptime", getUptime()},
		{"Load", getLoadAvg()},
		{"RAM", getMemoryInfo()},
	}
	for _, e := range info {
		sr.Header += fmt.Sprintf("%s: %s\n", utils.Wrap(e.name, c.padL, c.padR), e.content)
	}
}

// runCmd executes command and returns stdout as string
func runCmd(name string, args string, buf *bytes.Buffer) (string, error) {
	var retStr string
	cmd := exec.Command(name, args)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		retStr = utils.Warn("unavailable")
	} else {
		retStr = buf.String()
	}
	buf.Reset()
	return retStr, err
}

func getDistroName() (retStr string) {
	file, err := os.Open("/etc/unraid-version")
	if err != nil {
		retStr = utils.Warn("unavailable")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Look for pretty name
	re := regexp.MustCompile(`version=(.*)`)
	for scanner.Scan() {
		m := re.FindSubmatch(scanner.Bytes())
		if len(m) > 1 {
			// Remove quotes
			retStr = strings.Replace(string(m[1]), `"`, "", 2)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		retStr = utils.Warn("unavailable")
		return
	}
	return
}

func getUptime() string {
	var buf bytes.Buffer
	uptime, err := runCmd("uptime", "-p", &buf)
	if err != nil {
		return uptime
	}
	re := regexp.MustCompile(`(up\s|\n)`)
	return re.ReplaceAllString(uptime, "")
}

func getLoadAvg() string {
	loadavg, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return utils.Warn("unavailable")
	}
	var loadArr = strings.Split(string(loadavg), " ")
	return fmt.Sprintf("%s [1m], %s [5m], %s [15m]", loadArr[0], loadArr[1], loadArr[2])
}

func getMemoryInfo() (retStr string) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		retStr = utils.Warn("unavailable")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Look for active and total
	var memActive float64
	var memTotal float64
	reActive := regexp.MustCompile(`Active:\s+(\d+)`)
	reTotal := regexp.MustCompile(`MemTotal:\s+(\d+)`)
	for scanner.Scan() {
		if memTotal != 0 && memActive != 0 {
			break
		}
		if memActive == 0 {
			// Look for active
			m := reActive.FindSubmatch(scanner.Bytes())
			if len(m) > 1 {
				// Store as int
				memActive, _ = strconv.ParseFloat(string(m[1]), 64)
			}
		}
		if memTotal == 0 {
			m := reTotal.FindSubmatch(scanner.Bytes())
			if len(m) > 1 {
				// Store as int
				memTotal, _ = strconv.ParseFloat(string(m[1]), 64)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		retStr = utils.Warn("unavailable")
		return
	}
	// Convert to GB, meminfo is in kB
	return fmt.Sprintf("%.2f GB active of %.2f GB", memActive/1e6, memTotal/1e6)
}

func getKernel() string {
	var buf bytes.Buffer
	var kernel, _ = runCmd("uname", "-sr", &buf)
	return strings.ReplaceAll(kernel, "\n", "")
}
