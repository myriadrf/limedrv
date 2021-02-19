package limedrv

import "C"
import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/myriadrf/limedrv/limewrap"
	"github.com/racerxdl/fastconvert"
	"runtime"
	"strings"
	"unsafe"
)

const floatSize = 4
const int16Size = 2
const samplesWait = 100

func cleanString(s string) string {
	return strings.Trim(s, "\u0000 ")
}

type channelMessage struct {
	channel   int
	data      []complex64
	data16    []int16
	timestamp uint64
}

func FastI16BufferIQConvert(data []byte) []complex64 {
	var i16samples = len(data) / 2
	var out = make([]complex64, i16samples/2) // Each complex is 2 i16
	var pos = 0
	var itemsToRead = i16samples / 2

	for idx := 0; idx < itemsToRead; idx++ {
		var r = int16(binary.LittleEndian.Uint16(data[pos : pos+2]))
		var i = int16(binary.LittleEndian.Uint16(data[pos+2 : pos+4]))
		out[idx] = complex(float32(r)/32768, float32(i)/32768)
		pos += 4
	}

	return out
}

func ConvertC64toI16(dst []int16, src []complex64) {
	samples := len(src)
	if len(dst)/2 < samples {
		samples = len(dst) / 2
	}

	for idx := 0; idx < samples; idx++ {
		c := src[idx]
		dst[idx*2+0] = int16(real(c) * 32768)
		dst[idx*2+1] = int16(imag(c) * 32768)
	}
}

func streamTXLoop(ctx context.Context, channel LMSChannel, txCb func([]complex64, int)) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("got panic recovered: ", r)
		}
	}()
	//fmt.Fprintf(os.Stderr,"Worker Started")
	running := true
	sampleLength := floatSize
	if channel.parent.IQFormat == FormatInt16 || channel.parent.IQFormat == FormatInt12 {
		sampleLength = int16Size
	}

	var buffPtr uintptr
	var buff interface{}

	if sampleLength == int16Size {
		i16buff := make([]int16, fifoSize*2)
		buff = i16buff
		buffPtr = uintptr(unsafe.Pointer(&i16buff[0]))
	} else {
		c64buff := make([]complex64, fifoSize)
		buff = c64buff
		buffPtr = uintptr(unsafe.Pointer(&c64buff[0]))
	}

	rxData := make([]complex64, fifoSize)

	m := limewrap.NewLms_stream_meta_t()
	m.SetTimestamp(0)
	m.SetFlushPartialPacket(false)
	m.SetWaitForTimestamp(false)
	//fmt.Fprintf(os.Stderr,"Worker Running")
	for running {
		select {
		case _ = <-ctx.Done():
			//fmt.Fprintf(os.Stderr,"Worker Received stop", b)
			running = false
			return
		default:
		}
		if txCb != nil {
			txCb(rxData, channel.parentIndex) // Fill buffer
		}

		if sampleLength == floatSize {
			copy(buff.([]complex64), rxData)
		} else {
			ConvertC64toI16(buff.([]int16), rxData)
		}

		runtime.LockOSThread()
		sentSamples := limewrap.LMS_SendStream(channel.stream, buffPtr, fifoSize, m, samplesWait)
		runtime.UnlockOSThread()

		if sentSamples != fifoSize {
			fmt.Printf("Error sending samples. Expected %d sent %d\n", fifoSize, sentSamples)
		}
		runtime.Gosched()
	}
}

func streamRXLoop(ctx context.Context, c chan<- channelMessage, channel LMSChannel, send16 bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("got panic recovered: ", r)
		}
	}()
	//fmt.Fprintf(os.Stderr,"Worker Started")
	running := true
	sampleLength := floatSize
	if channel.parent.IQFormat == FormatInt16 || channel.parent.IQFormat == FormatInt12 {
		sampleLength = int16Size
	}
	buff := make([]byte, fifoSize*sampleLength*2) // 16k IQ samples
	buffPtr := uintptr(unsafe.Pointer(&buff[0]))

	m := limewrap.NewLms_stream_meta_t()
	m.SetTimestamp(0)
	m.SetFlushPartialPacket(false)
	m.SetWaitForTimestamp(false)
	//fmt.Println("Worker Running")
	for running {
		select {
		case _ = <-ctx.Done():
			//fmt.Println("Worker Received stop")
			running = false
			return
		default:
		}
		runtime.LockOSThread()
		recvSamples := limewrap.LMS_RecvStream(channel.stream, buffPtr, fifoSize, m, samplesWait)
		runtime.UnlockOSThread()

		if recvSamples > 0 {
			chunk := buff[:sampleLength*recvSamples*2]
			cm := channelMessage{
				channel:   channel.parentIndex,
				timestamp: m.GetTimestamp(),
			}

			if sampleLength == floatSize {
				// Float32
				cm.data = fastconvert.ByteArrayToComplex64Array(chunk)
			} else {
				// Int16
				if send16 {
					cm.data16 = fastconvert.ByteArrayToInt16LEArray(chunk)
				} else {
					cm.data = FastI16BufferIQConvert(chunk)
				}
			}

			c <- cm
		} else if recvSamples == -1 {
			fmt.Printf("Error receiving samples from channel %d\n", channel.parentIndex)
		}
		runtime.Gosched()
	}
	//fmt.Println("Worker stopped")
}

func createLms_range_t() limewrap.Lms_range_t {
	return limewrap.NewLms_range_t()
}

func createLms_stream_t() limewrap.Lms_stream_t {
	return limewrap.NewLms_stream_t()
}

func idev2dev(deviceinfo i_deviceinfo) DeviceInfo {
	var deviceStr = string(deviceinfo.DeviceName[:64])
	var z = strings.Split(deviceStr, ",")

	var DeviceName string
	var Media string
	var Module string
	var Addr string
	var Serial string

	for i := 0; i < len(z); i++ {
		var k = strings.Split(z[i], "=")
		if len(k) == 1 {
			DeviceName = k[0]
		} else {
			switch strings.ToLower(strings.Trim(k[0], " ")) {
			case "media":
				Media = cleanString(k[1])
				break
			case "module":
				Module = cleanString(k[1])
				break
			case "addr":
				Addr = cleanString(k[1])
				break
			case "serial":
				Serial = cleanString(k[1])
				break
			}
		}
	}

	return DeviceInfo{
		DeviceName:          DeviceName,
		Media:               Media,
		Module:              Module,
		Addr:                Addr,
		Serial:              Serial,
		FirmwareVersion:     cleanString(string(deviceinfo.FirmwareVersion[:16])),
		HardwareVersion:     cleanString(string(deviceinfo.HardwareVersion[:16])),
		GatewareVersion:     cleanString(string(deviceinfo.GatewareVersion[:16])),
		GatewareTargetBoard: cleanString(string(deviceinfo.GatewareTargetBoard[:16])),
		origDevInfo:         deviceinfo,
	}
}
