package player

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/schollz/pianoai/ai2"
	"github.com/schollz/pianoai/music"
	"github.com/schollz/pianoai/piano"
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

// Player is the main structure which facilitates the Piano, and the AI.
// The Player spawns threads for listening to events on the Piano, and also
// spawns threads for playing notes on the piano. It also spawns threads
// for doing the machine learning and using the results.
type Player struct {
	// BPM is the beats per minute
	BPM int
	// Beat counts the number of 1/64 beats
	Tick int
	// Key stores the key of the song (TODO: Add in key-signature constraints)
	Key string

	// Piano is the piano that does the playing, the MIDI keyboard
	Piano *piano.Piano
	// MusicFuture is a map of future chords to play
	MusicFuture *music.Music
	// MusicHistory is a map of all the previous notes played
	MusicHistory     *music.Music
	MusicHistoryFile string

	// AI stores the AI being used
	AI *ai2.AI
	// BeatsOfSilence waits this number of beats before asking
	// the AI for an improvisation
	BeatsOfSilence int
	// lastNote is the beat of the last note played
	lastNote int
	// HighPassFilter only uses notes above a certain level
	// for computing last note
	HighPassFilter int
	// KeysCurrentlyPressed keeps track of whether a key is down (should be 0 if no keys are down)
	KeysCurrentlyPressed int

	// Listening frequency (to determine tick size)
	ListeningRateHertz int
	// Number of ticks per beat
	TicksPerBeat int
	// 1/Quantize = shortest possible note
	Quantize int

	// flag to allow only manually activation
	ManualAI bool

	// UseHostVelocity changes emitted notes to follow the velocity of the host
	UseHostVelocity bool

	LastHostPress int
	IsImprovising bool
	lastVelocity  int
}

// New initializes the parameters and connects up the piano
func New(bpm, listenHertz int, debug bool) (p *Player, err error) {
	p = new(Player)
	logger := log.WithFields(log.Fields{
		"function": "Player.Init",
	})
	if !debug {
		log.SetLevel(log.InfoLevel)
	}
	p.BPM = bpm
	p.Tick = 0
	p.Key = "C"
	p.Quantize = 64

	logger.Debug("Loading piano")
	p.Piano, err = piano.New()
	if err != nil {
		return
	}

	logger.Debug("Loading music")
	p.MusicFuture = music.New()
	var errOpening error
	p.MusicHistoryFile = "music_history.json"
	p.MusicHistory, errOpening = music.Open(p.MusicHistoryFile)
	if errOpening != nil {
		logger.Warn(errOpening.Error())
		p.MusicHistory = music.New()
	} else {
		logger.Info("Loaded previous music history")
	}

	logger.Debug("Loading AI")
	p.ListeningRateHertz = listenHertz
	p.BeatsOfSilence = 2
	p.HighPassFilter = 65
	p.lastNote = 0

	p.TicksPerBeat = int(float64(p.ListeningRateHertz) / (float64(p.BPM) / 60))

	p.AI = ai2.New(p.TicksPerBeat)
	p.AI.HighPassFilter = p.HighPassFilter

	return
}

// Close will do the shutdown routines before exiting
func (p *Player) Close() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Player.Close",
	})
	logger.Debug("Closing piano...")
	err = p.Piano.Close()
	if err != nil {
		logger.Error(err.Error())
	}
	return
}

// Start initializes the metronome which keeps track of beats
// Each beat will start new threads to Emit new chords, and/or
// generate new Improvisation
func (p *Player) Start() {
	logger := log.WithFields(log.Fields{
		"function": "Player.Start",
	})

	// Exit on Ctl+C
	doneChan := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.Debugf("%+v", sig)
			// sig is a ^C, handle it
			p.Close()
			doneChan <- true
		}
	}()

	// start listening
	go p.Listen()

	p.Tick = 0
	tickTime := 1000 * time.Duration(1000000/p.ListeningRateHertz)
	tickChan := time.NewTicker(tickTime).C
	logger.Infof("BPM:  %d, tick size: %s (%d ticks / beat)", p.BPM, tickTime.String(), p.TicksPerBeat)
	for {
		select {
		case <-tickChan:
			// if p.Tick == math.Trunc(p.Tick) {
			// 	logger.Debugf("beat %2.0f", p.Tick)
			// }
			p.Tick += 1
			go p.Emit(p.Tick)

			if !p.ManualAI {
				if p.Tick-p.lastNote > (p.TicksPerBeat*p.BeatsOfSilence) && p.KeysCurrentlyPressed == 0 && !p.AI.IsLearning {
					logger.Info("Silence exceeded, trying to improvise")
					p.lastNote = p.Tick
					go p.Improvisation()
				}
			}

			// if math.Mod(float64(p.Tick), 64) == 0 {
			// 	logger.WithFields(log.Fields{
			// 		"Beat":     p.Tick,
			// 		"LastNote": p.lastNote,
			// 		"KeysDown": p.KeysCurrentlyPressed,
			// 	}).Debug("metronome")
			// }

		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}

func (p *Player) Teach() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Player.Teach",
	})
	logger.Info("Sending history to AI")
	err = p.AI.Learn(p.MusicHistory)
	if err != nil {
		logger.Warn(err.Error())
		return
	}
	return
}

// Improvisation generates an improvisation from the AI
// and loads into the next beats to be playing
func (p *Player) Improvisation() {
	logger := log.WithFields(log.Fields{
		"function": "Player.Improvisation",
	})
	if p.MusicFuture.HasFuture(p.Tick) || p.IsImprovising {
		logger.Debug("Improvising is already in progress")
		return
	}
	p.IsImprovising = true
	err := p.Teach()
	if err != nil {
		return
	}
	logger.Info("Getting improvisation")
	notes, err := p.AI.Lick(p.Tick)
	if err != nil {
		logger.Error(err.Error())
	}
	newNotes := notes.GetAll()
	for _, note := range newNotes {
		p.MusicFuture.AddNote(note)
	}
	logger.Infof("Added %d notes from AI", len(newNotes))
	p.IsImprovising = false
}

// Emit will play/stop notes depending on the current beat.
// This should be run in a separate thread.
func (p *Player) Emit(beat int) {
	hasNotes, notes := p.MusicFuture.Get(beat)
	if hasNotes {
		if p.Tick-p.LastHostPress > p.BeatsOfSilence*p.TicksPerBeat && p.KeysCurrentlyPressed == 0 {
			if p.UseHostVelocity && p.lastVelocity > 0 {
				for i := range notes {
					notes[i].Velocity = p.lastVelocity
				}
			}
			go p.Piano.PlayNotes(notes, p.BPM)
		}
		p.lastNote = p.Tick
	}
}

// Listen tells the player to listen to events from the
// piano MIDI connection. This is meant to be run in a
// separate thread.
func (p *Player) Listen() {
	logger := log.WithFields(log.Fields{
		"function": "Player.Listen",
	})

	ch := p.Piano.InputStream.Listen()
	prevTick := p.Tick
	for {
		event := <-ch
		tickOfNote := p.Tick
		// only allow up to 64th notes
		if tickOfNote-prevTick < p.TicksPerBeat/p.Quantize {
			tickOfNote = prevTick
		}
		note := music.Note{
			On:       event.Data2 > 0,
			Pitch:    int(event.Data1),
			Velocity: int(event.Data2),
			Beat:     tickOfNote,
		}
		prevTick = tickOfNote

		if note.Pitch == 21 {
			if !note.On {
				continue
			}
			p.MusicHistory.Save(p.MusicHistoryFile)
			logger.Infof("Saved %s.json", p.MusicHistoryFile)
		} else if note.Pitch == 22 {
			if !note.On {
				continue
			}
			logger.Info("Playing back history")
			for _, note := range p.MusicHistory.GetAll() {
				logger.Infof("Adding %+v to future", note)
				p.MusicFuture.AddNote(note)
			}
			p.Tick = 0
		} else if note.Pitch == 107 {
			if !note.On {
				continue
			}
			p.Teach()
		} else if note.Pitch == 108 {
			if !note.On {
				continue
			}
			p.Improvisation()
		} else {
			if !note.On && note.Pitch > p.HighPassFilter {
				p.lastNote = p.Tick
				p.KeysCurrentlyPressed--
			}
			if note.On && note.Pitch > p.HighPassFilter {
				p.LastHostPress = p.Tick
				p.KeysCurrentlyPressed++
			}
			if note.On && p.UseHostVelocity {
				p.lastVelocity = note.Velocity
			}
			logger.Infof("Adding %+v", note)
			go p.MusicHistory.AddNote(note)
		}
	}
}
