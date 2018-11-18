package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gorilla/websocket"
	"github.com/schollz/midi"
	"github.com/schollz/pianoai/src/rtmidi"
)

var mainTemplate *template.Template
var numConnected int
var noteChannel chan Note

type Note struct {
	Name     string `json:"name"`
	Midi     uint8  `json:"midi"`
	Velocity uint8  `json:"velocity"`
}

func (note Note) String() string {
	on := "on"
	if note.Velocity == 0 {
		on = "off"
	}
	return fmt.Sprintf("%s %s(%d) %d", on, note.Name, note.Midi, note.Velocity)
}

type TemplateRender struct {
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	var err error

	// b, err := Asset("assets/index.html")
	b, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		panic(err)
	}
	mainTemplate = template.Must(template.New("main").Parse(string(b)))

	// setup note channel
	noteChannel = make(chan Note, 1024)
}

var Version string

func main() {
	var err error
	var debug = flag.Bool("debug", false, "debug mode")
	var showVersion = flag.Bool("v", false, "show version")
	flag.Parse()

	if *debug {
		err = setLogLevel("debug")
	} else {
		err = setLogLevel("info")
	}
	defer log.Flush()

	if *showVersion {
		fmt.Println(Version)
		return
	}

	cSignal := make(chan os.Signal)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		log.Debug("cleanup")
		log.Flush()
		os.Exit(1)
	}()

	f, err := os.Create("midi.mid")
	if err != nil {
		log.Error(err)
	}
	defer func() {
		f.Close()
	}()
	e := midi.NewEncoder(f, midi.SingleTrack, 96)
	tr := e.NewTrack()

	// 1 beat with 1 note for nothing
	tr.Add(1, midi.NoteOff(0, 60))

	vel := 90
	//C3 to B3
	var j float64
	for i := 60; i < 72; i++ {
		tr.Add(j, midi.NoteOn(0, i, vel))
		tr.Add(1, midi.NoteOff(0, i))
		j = 1
	}
	tr.Add(1, midi.EndOfTrack())

	if err := e.Write(); err != nil {
		log.Error(err)
	}

	log.Debug(rtmidi.CompiledAPI())

	in, err := rtmidi.NewMIDIInDefault()
	if err != nil {
		log.Error(err)
	}
	defer in.Close()

	n, err := in.PortCount()
	if err != nil {
		log.Error(err)
	}
	for i := 0; i < n; i++ {
		name, err := in.PortName(i)
		if err == nil {
			log.Debug(name)
			if err := in.OpenPort(i, "RtMidi"); err != nil {
				log.Debug(err)
			}
		} else {
			log.Debug(err)
		}
	}

	in.SetCallback(func(m rtmidi.MIDIIn, msg []byte, t float64) {
		// log.Debug(msg, t)
		if msg[0] == 128 {
			msg[2] = 0
		}
		note := Note{midiToNote(msg[1]), msg[1], msg[2]}
		log.Debugf("%s", note)
		if numConnected > 0 {
			noteChannel <- note
		}
	})

	// for {
	// 	m, t, err := in.Message()
	// 	if len(m) > 0 {
	// 		log.Debug(m, t, err)
	// 	}
	// }

	err = serve()
	if err != nil {
		log.Error(err)
	}
}

func serve() (err error) {
	log.Info("running on port 8152")
	http.HandleFunc("/", handler)
	return http.ListenAndServe(":8152", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%v %v %v %s", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

func handle(w http.ResponseWriter, r *http.Request) (err error) {
	// very special paths
	if r.URL.Path == "/robots.txt" {
		// special path
		_, err = w.Write([]byte(`User-agent: * 
Disallow: /`))
		return
	} else if r.URL.Path == "/favicon.ico" {
		// TODO
	} else if r.URL.Path == "/sitemap.xml" {
		// TODO
	} else if strings.HasPrefix(r.URL.Path, "/static") {
		// special path /static
		return handleStatic(w, r)
	} else if strings.HasSuffix(r.URL.Path, "/ws") {
		return handleWebsocket(w, r)
	}
	return handleMain(w, r)
}

func handleStatic(w http.ResponseWriter, r *http.Request) (err error) {
	page := r.URL.Path
	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	w.Header().Set("Content-Encoding", "gzip")
	if strings.HasPrefix(page, "/static") {
		page = "assets/" + strings.TrimPrefix(page, "/static/")
		b, _ := Asset(page + ".gz")
		if strings.Contains(page, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		} else if strings.Contains(page, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.Contains(page, ".png") {
			w.Header().Set("Content-Type", "image/png")
		} else if strings.Contains(page, ".json") {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write(b)
	}

	return
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	// handle websockets on this page
	c, errUpgrade := wsupgrader.Upgrade(w, r, nil)
	if errUpgrade != nil {
		return errUpgrade
	}
	defer c.Close()
	numConnected += 1
	defer func() {
		numConnected -= 1
	}()

	type WebsocketMessage struct {
		Message string `json:"message"`
		Note    Note   `json:"note,omitempty"`
	}

	go func() {
		var wm WebsocketMessage
		for {
			errRead := c.ReadJSON(&wm)
			if errRead != nil {
				log.Debug(errRead)
				return
			}
			switch wm.Message {
			case "ping":
				c.WriteJSON(WebsocketMessage{Message: "pong"})
			}
		}
	}()

	for {
		note := <-noteChannel
		err = c.WriteJSON(WebsocketMessage{"note", note})
		if err != nil {
			log.Debug(err)
			return
		}
	}
	return
}

func handleMain(w http.ResponseWriter, r *http.Request) (err error) {
	tr := TemplateRender{}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	return mainTemplate.Execute(gz, tr)
}

// setLogLevel determines the log level
func setLogLevel(level string) (err error) {

	// https://en.wikipedia.org/wiki/ANSI_escape_code#3/4_bit
	// https://github.com/cihub/seelog/wiki/Log-levels
	appConfig := `
	<seelog minlevel="` + level + `">
	<outputs formatid="stdout">
	<filter levels="debug,trace">
		<console formatid="debug"/>
	</filter>
	<filter levels="info">
		<console formatid="info"/>
	</filter>
	<filter levels="critical,error">
		<console formatid="error"/>
	</filter>
	<filter levels="warn">
		<console formatid="warn"/>
	</filter>
	</outputs>
	<formats>
		<format id="stdout"   format="%Date %Time [%LEVEL] %File %FuncShort:%Line %Msg %n" />
		<format id="debug"   format="%Date %Time %EscM(37)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="info"    format="%Date %Time %EscM(36)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="warn"    format="%Date %Time %EscM(33)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="error"   format="%Date %Time %EscM(31)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
	</formats>
	</seelog>
	`
	logger, err := log.LoggerFromConfigAsBytes([]byte(appConfig))
	if err != nil {
		return
	}
	log.ReplaceLogger(logger)
	return
}

var chromatic = []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

func midiToNote(midiNum uint8) string {
	midiNumF := float64(midiNum)
	return fmt.Sprintf("%s%1.0f", chromatic[int(math.Mod(midiNumF, 12))], math.Floor(midiNumF/12.0-1))
}
