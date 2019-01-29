package main

import (
    "fmt"
    "os"

    "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

func fatalOnErr(err error) {
    if err != nil {
        fmt.Printf("Fatal error: %s\n", err.Error())
        os.Exit(1)
    }
}

func main() {
    nvml.Init()
    defer nvml.Shutdown()

    count, err := nvml.GetDeviceCount()
    fatalOnErr(err)
    fmt.Printf("Found %d devices\n", count)

    for id := uint(0); id < count; id++ {
        device, err := nvml.NewDevice(id)
        fatalOnErr(err)

        info, err := device.GetAllRunningProcesses()
        fatalOnErr(err)

        if len(info) == 0 {
            fmt.Printf("Device %d is idle ...\n", id)
        } else {
            for i := range info {
                ii := info[i]
                fmt.Printf("Device %d running %v [%v] type=%v mem=%v\n",
                            id, ii.Name, ii.PID, ii.Type, ii.MemoryUsed)
            }
        }
    }
    fmt.Printf("Done...\n")
}


