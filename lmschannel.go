package limedrv

import (
	"fmt"
	"github.com/myriadrf/limedrv/limewrap"
)

// LMSChannel is the struct that represents a Channel from a LMSDevice.
// It can be either a RX or TX Channel, defined by the field IsRX.
// It also contains the list of available antenna ports.
type LMSChannel struct {
	Antennas []LMSAntenna
	IsRX     bool

	parent                  *LMSDevice
	parentIndex             int
	stream                  limewrap.Lms_stream_t
	currentDigitalBandwidth float64
	digitalFilterEnabled    bool
	advancedFiltering       bool
}

// Enable enables this channel from the read / write callback
func (c *LMSChannel) Enable() *LMSChannel {
	c.parent.EnableChannel(c.parentIndex, c.IsRX)
	return c
}

// Disable disables this channel from the read / write callback
func (c *LMSChannel) Disable() *LMSChannel {
	c.parent.DisableChannel(c.parentIndex, c.IsRX)
	return c
}

// SetGainDB sets this channel gain in decibels
func (c *LMSChannel) SetGainDB(gain uint) *LMSChannel {
	c.parent.SetGainDB(c.parentIndex, c.IsRX, gain)
	return c
}

// SetGainNormalized sets the channel normalized gain. [0-1]
func (c *LMSChannel) SetGainNormalized(gain float64) *LMSChannel {
	c.parent.SetGainNormalized(c.parentIndex, c.IsRX, gain)
	return c
}

// GetGainDB returns the channel current gain in decibels
func (c *LMSChannel) GetGainDB() uint {
	return c.parent.GetGainDB(c.parentIndex, c.IsRX)
}

// GetGainNormalized returns the channel current normalized gain. [0-1]
func (c *LMSChannel) GetGainNormalized() float64 {
	return c.parent.GetGainNormalized(c.parentIndex, c.IsRX)
}

// SetLPF sets the Analog Low Pass filter bandwidth for the current channel.
func (c *LMSChannel) SetLPF(bandwidth float64) *LMSChannel {
	c.parent.SetLPF(c.parentIndex, c.IsRX, bandwidth)
	return c
}

// GetLPF gets the current Analog Low Pass filter bandwidth for the current channel.
func (c *LMSChannel) GetLPF() float64 {
	return c.parent.GetLPF(c.parentIndex, c.IsRX)
}

// EnableLPF enables the Analog Low Pass filter for the current channel.
func (c *LMSChannel) EnableLPF() *LMSChannel {
	c.parent.EnableLPF(c.parentIndex, c.IsRX)
	return c
}

// DisableLPF disables the Analog Low Pass filter for the current channel.
func (c *LMSChannel) DisableLPF() *LMSChannel {
	c.parent.EnableLPF(c.parentIndex, c.IsRX)
	return c
}

// SetDigitalLPF sets the current channel digital filter (GFIR) to low pass with specified bandwidth.
func (c *LMSChannel) SetDigitalLPF(bandwidth float64) *LMSChannel {
	c.parent.SetDigitalFilter(c.parentIndex, c.IsRX, bandwidth)
	return c
}

// EnableDigitalLPF enables current channel digital filter (GFIR)
func (c *LMSChannel) EnableDigitalLPF() *LMSChannel {
	c.parent.EnableDigitalFilter(c.parentIndex, c.IsRX)
	return c
}

// DisableDigitalLPF disables current channel digital filter (GFIR)
func (c *LMSChannel) DisableDigitalLPF() *LMSChannel {
	c.parent.DisableDigitalFilter(c.parentIndex, c.IsRX)
	return c
}

// SetAntenna sets the current channel antenna port
func (c *LMSChannel) SetAntenna(idx int) *LMSChannel {
	c.parent.SetAntenna(idx, c.parentIndex, c.IsRX)
	return c
}

// SetAntennaByName sets the current channel antenna port by name.
// Example: LNAW
func (c *LMSChannel) SetAntennaByName(name string) *LMSChannel {
	c.parent.SetAntennaByName(name, c.parentIndex, c.IsRX)
	return c
}

// SetCenterFrequency sets the current channel center frequency in hertz.
func (c *LMSChannel) SetCenterFrequency(centerFrequency float64) *LMSChannel {
	c.parent.SetCenterFrequency(c.parentIndex, c.IsRX, centerFrequency)
	return c
}

// GetCenterFrequency returns the current channel center frequency in hertz.
func (c *LMSChannel) GetCenterFrequency() float64 {
	return c.parent.GetCenterFrequency(c.parentIndex, c.IsRX)
}

// String returns a representation of the channel
func (c *LMSChannel) String() string {
	var str = fmt.Sprintf("\nIs RX: %t\nAntennas: %d", c.IsRX, len(c.Antennas))
	for i := 0; i < len(c.Antennas); i++ {
		str = fmt.Sprintf("%s\n\t%s", str, c.Antennas[i].String())
	}

	return str
}

func (c *LMSChannel) start() {
	if c.stream != nil {
		limewrap.LMS_Calibrate(c.parent.dev, !c.IsRX, int64(c.parentIndex), c.GetLPF(), 0)
		limewrap.LMS_StartStream(c.stream)
	} else {
		panic("request to start a nil stream")
	}
}

func (c *LMSChannel) stop() {
	if c.stream != nil {
		limewrap.LMS_StopStream(c.stream)
	}
}
