package main

import (
	"github.com/myriadrf/limedrv"
	"log"
	"os"
	"time"
)

func OnSamples(data []int16, channel int, timestamp uint64) {
	log.Println("Received samples from channel", channel, "with timestamp", timestamp)
}

func NeedSamples(data []complex64, channel int) {
	log.Printf("Channel %d needs %d samples.", channel, len(data))
	// TODO: Put something in data
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

	d.SetI16CallbackMode(true)

	//d.EnableChannel(limedrv.ChannelA, true)
	//d.EnableChannel(limedrv.ChannelB, true)
	//d.SetAntennaByName("LNAW", limedrv.ChannelA, true)
	//d.SetAntennaByName("LNAW", limedrv.ChannelB, true)

	var rxCh = d.RXChannels[limedrv.ChannelA]
	var txCh = d.TXChannels[limedrv.ChannelA]

	rxCh.Enable().
		SetAntennaByName(limedrv.LNAW).
		SetGainNormalized(0.5).
		SetLPF(1e6).
		EnableLPF().
		SetCenterFrequency(106.3e6)

	txCh.Enable().
		SetAntennaByName(limedrv.BAND1).
		SetGainNormalized(0.5).
		SetLPF(1e6).
		EnableLPF().
		SetCenterFrequency(120e6)

	d.SetI16Callback(OnSamples)
	d.SetTXCallback(NeedSamples)

	d.Start()

	time.Sleep(5 * 1000 * time.Millisecond)

	d.Stop()

	log.Println("Closing")
	d.Close()

	log.Println("Closed!")
}
