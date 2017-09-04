package piano

import (
	"sync"

	"github.com/rakyll/portmidi"
	"github.com/schollz/pianoai/music"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

// Piano is the AI class for the piano
type Piano struct {
	InputDevice  portmidi.DeviceID
	OutputDevice portmidi.DeviceID
	outputStream *portmidi.Stream
	InputStream  *portmidi.Stream
	sync.Mutex
}

// New sets the device ports. Optionally you can
// pass the input and output ports, respectively.
func New(ports ...int) (p *Piano, err error) {
	p = new(Piano)
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
		logger.Debugf("%d) %s %s %s", i, deviceInfo.Interface, deviceInfo.Name, inputOutput)
	}
	if len(ports) == 2 {
		p.InputDevice = portmidi.DeviceID(ports[0])
		p.OutputDevice = portmidi.DeviceID(ports[1])
	}
	logger.Infof("Using input device %d and output device %d", p.InputDevice, p.OutputDevice)

	logger.Debug("Opening output stream")
	p.outputStream, err = portmidi.NewOutputStream(p.OutputDevice, 1024, 0)
	if err != nil {
		if err != nil {
			logger.WithFields(log.Fields{
				"msg": "problem getting output stream from device " + string(p.OutputDevice),
			}).Error(err.Error())
			return
		}

	}

	logger.Debug("Opening input stream")
	p.InputStream, err = portmidi.NewInputStream(p.InputDevice, 1024)
	if err != nil {
		if err != nil {
			logger.WithFields(log.Fields{
				"msg": "problem getting input stream from device " + string(p.InputDevice),
			}).Error(err.Error())
			return
		}

	}
	return
}

// Close will shutdown the streams
// and gracefully terminate.
func (p *Piano) Close() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.Close",
	})
	logger.Debug("Closing output stream")
	p.outputStream.Close()
	logger.Debug("Closing input stream")
	p.InputStream.Close()
	logger.Debug("Terminating portmidi")
	portmidi.Terminate()
	return
}

// PlayNotes will play all the notes
func (p *Piano) PlayNotes(notes []music.Note, bpm int) (err error) {
	p.Lock()
	defer p.Unlock()
	logger := log.WithFields(log.Fields{
		"function": "Piano.PlayNotes",
	})
	for _, note := range notes {
		if note.On {
			logger.WithFields(log.Fields{
				"p": note.Pitch,
				"v": note.Velocity,
			}).Debugf("on, beat %d", note.Beat)
			err = p.outputStream.WriteShort(0x90, int64(note.Pitch), int64(note.Velocity))
			if err != nil {
				logger.WithFields(log.Fields{
					"p":   note.Pitch,
					"v":   note.Velocity,
					"msg": "problem turning on",
				}).Error(err.Error())
				return
			}
		} else {
			logger.WithFields(log.Fields{
				"p": note.Pitch,
				"v": note.Velocity,
			}).Debugf("off, beat %d", note.Beat)
			err = p.outputStream.WriteShort(0x80, int64(note.Pitch), int64(note.Velocity))
			if err != nil {
				logger.WithFields(log.Fields{
					"p":   note.Pitch,
					"v":   note.Velocity,
					"msg": "problem turning off",
				}).Error(err.Error())
				return
			}
		}
	}
	return
}
