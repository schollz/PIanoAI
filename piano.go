package main

import (
	"time"

	"github.com/apex/log"
	"github.com/schollz/portmidi"
)

// Piano is the AI class for the piano
type Piano struct {
	InputDevice  portmidi.DeviceID
	OutputDevice portmidi.DeviceID
}

// Init sets the device ports. Optionally you can
// pass the input and output ports, respectively.
func (p *Piano) Init(ports ...int) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.Init",
	})
	logger.Debug("Initializing portmidi...")
	err = portmidi.Initialize()
	if err != nil {
		logger.WithFields(log.Fields{
			"msg": "initiailization failed",
		}).Error(err.Error())
		return
	}
	numDevices := portmidi.CountDevices()
	logger.Debugf("Found %d devices", numDevices)
	for i := 0; i < numDevices; i++ {
		deviceInfo := portmidi.Info(portmidi.DeviceID(i))
		var inputOutput string
		if deviceInfo.IsOutputAvailable {
			inputOutput = "output"
			p.OutputDevice = portmidi.DeviceID(i)
		} else {
			inputOutput = "input"
			p.InputDevice = portmidi.DeviceID(i)
		}
		logger.Infof("%d) %s %s %s", i, deviceInfo.Interface, deviceInfo.Name, inputOutput)
	}
	if len(ports) == 2 {
		p.InputDevice = portmidi.DeviceID(ports[0])
		p.OutputDevice = portmidi.DeviceID(ports[1])
	}
	logger.Infof("Using input device %d and output device %d", p.InputDevice, p.OutputDevice)
	return
}

func (p *Piano) PlayNotes() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.PlayNotes",
	})
	out, err := portmidi.NewOutputStream(p.OutputDevice, 1024, 0)
	if err != nil {
		logger.WithFields(log.Fields{
			"msg": "output stream failed to connect",
		}).Error(err.Error())
		return
	}
	logger.Info("Playing notes..")
	// note on events to play C major chord
	out.WriteShort(0x90, 60, 100)
	out.WriteShort(0x90, 64, 100)
	out.WriteShort(0x90, 67, 100)

	// notes will be sustained for 2 seconds
	time.Sleep(2 * time.Second)

	// note off events
	out.WriteShort(0x80, 60, 100)
	out.WriteShort(0x80, 64, 100)
	out.WriteShort(0x80, 67, 100)

	out.Close()
	return
}
