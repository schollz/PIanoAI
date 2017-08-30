package main

import (
	"os"
	"time"

	"github.com/schollz/portmidi"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	var piano Piano
	piano.Init()
	piano.PlayNotes()
}

// Piano is the AI class for the piano
type Piano struct {
	InputDevice  portmidi.DeviceID
	OutputDevice portmidi.DeviceID
}

func (p *Piano) Init() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.Init",
	})
	logger.Debug("Initializing portmidi...")
	err = portmidi.Initialize()
	if err != nil {
		logger.WithFields(log.Fields{
			"note": "initiailization failed",
		}).Error(err)
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
	logger.Infof("Using input device %d and output device %d", p.InputDevice, p.OutputDevice)
	return
}

func (p *Piano) PlayNotes() {
	logger := log.WithFields(log.Fields{
		"function": "Piano.PlayNotes",
	})
	out, err := portmidi.NewOutputStream(p.OutputDevice, 1024, 0)
	if err != nil {
		log.Fatal(err)
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
}
