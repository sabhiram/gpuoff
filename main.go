package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

// multiValueFlag implements the flag.Value interface for a list of strings
// that can be passed using `--foo bar --foo baz` to produce [bar, baz].
type multiValueFlag []string

// String returns the string representation of the multi value flag.
func (mvf *multiValueFlag) String() string {
	return "[" + strings.Join(*mvf, ", ") + "]"
}

// Set sets one part of the multi value flag.
func (mvf *multiValueFlag) Set(value string) error {
	*mvf = append(*mvf, value)
	return nil
}

var (
	zeroTime = time.Time{}
)

var (
	// ignores keeps track of all the processes that we need to ignore. The
	// given string is compiled into a regex before it is matched.
	ignores multiValueFlag

	// interval keeps track of how long we wait before checking for GPU status.
	interval time.Duration

	// timeout keeps track of how long the GPU can stay idle before a shutdown.
	timeout time.Duration
)

// fatalOnErr fatals on errors.
func fatalOnErr(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

// isIgnoredProcessName returns if the input string matches one of the ignored
// process regexes.
func isIgnoredProcessName(pn string) (bool, error) {
	for _, ignore := range ignores {
		matched, err := regexp.Match(ignore, []byte(pn))
		if err != nil {
			return false, err
		}
		if matched == true {
			return true, nil
		}
	}
	return false, nil
}

// isGPUIdle checks `count` number of devices to see if any of them are running
// non-ignored processes.
func isGPUIdle(count uint) (bool, error) {
	for id := uint(0); id < count; id++ {
		device, err := nvml.NewDevice(id)
		if err != nil {
			return false, err
		}

		info, err := device.GetAllRunningProcesses()
		if err != nil {
			return false, err
		}

		for i := range info {
			ignored, err := isIgnoredProcessName(info[i].Name)
			if err != nil {
				return false, err
			}
			if !ignored {
				return false, nil // Busy
			}
		}
	}
	return true, nil // Idle
}

func main() {
	nvml.Init()
	defer nvml.Shutdown()

	// Get the device count, this should not change once we are running.
	count, err := nvml.GetDeviceCount()
	fatalOnErr(err)

	// Once per interval ticker.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	idleTime := time.Time{}
	for {
		select {
		case <-ticker.C:
			idle, err := isGPUIdle(count)
			fatalOnErr(err)

			if idle && idleTime.IsZero() {
				println("GPU is now idle ...")
				idleTime = time.Now()
			} else if !idle && !idleTime.IsZero() {
				println("GPU is working ...")
				idleTime = zeroTime
			}

			if !idleTime.IsZero() {
				if time.Now().Sub(idleTime) > timeout {
					println("GPU idle timeout, shutting down ... \n", timeout)
					fatalOnErr(syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF))
					os.Exit(0)
				}
			}
		}
	}
}

func init() {
	flag.Var(&ignores, "ignore", "list of processes to ignore (always on like XORG)")
	flag.Var(&ignores, "i", "list of processes to ignore (always on like XORG) (short)")
	flag.DurationVar(&interval, "interval", 10*time.Second, "duration between checking GPU status")
	flag.DurationVar(&interval, "n", 10*time.Second, "duration between checking GPU status (short)")
	flag.DurationVar(&timeout, "timeout", 15*time.Minute, "duration to shutdown after GPU(s) are idle")
	flag.DurationVar(&timeout, "t", 15*time.Minute, "duration to shutdown after GPU(s) are idle (short)")
	flag.Parse()
}
