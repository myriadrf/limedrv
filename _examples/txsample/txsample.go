// Same as https://github.com/myriadrf/LimeSuite/blob/master/src/examples/basicTX.cpp
package main

import (
    "github.com/myriadrf/limedrv"
    "log"
    "math"
    "os"
    "time"
)

const frequency = 500e6
const sampleRate = 5e6
const toneFrequency = 1e6
const fRatio = toneFrequency / sampleRate

func NeedSamples(data []complex64, channel int) {
    for i := 0; i < len(data); i++ {
        w := 2 * math.Pi * fRatio * float64(i)
        data[i] = complex64(complex(math.Cos(w), math.Sin(w)))
    }
}

func main() {
    //profiler := profile.Start()
    //defer profiler.Stop()
    devices := limedrv.GetDevices()

    log.Printf("Found %d devices.\n", len(devices))

    if len(devices) == 0 {
        log.Println("No devices found.")
        os.Exit(1)
    }

    if len(devices) > 1 {
        log.Println("More than one device found. Selecting first one.")
    }

    var di = devices[0]

    log.Printf("Opening device %s\n", di.DeviceName)

    var d = limedrv.Open(di)
    log.Println("Opened!")

    log.Println(d.String())

    d.SetSampleRate(sampleRate, 4)

    var txCh = d.TXChannels[limedrv.ChannelB]

    txCh.Enable().
        SetAntennaByName(limedrv.BAND2).
        SetGainNormalized(0.7).
        SetCenterFrequency(frequency)

    d.SetTXCallback(NeedSamples)

    d.Start()

    time.Sleep(5 * 1000 * time.Millisecond)

    d.Stop()

    log.Println("Closing")
    d.Close()

    log.Println("Closed!")
}
