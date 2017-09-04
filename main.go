package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/pianoai/ai2"
	"github.com/schollz/pianoai/player"
	"github.com/urfave/cli"
)

var version string

func main() {

	app := cli.NewApp()
	app.Version = version
	app.Compiled = time.Now()
	app.Name = "pianoai"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "bpm",
			Value: 120,
			Usage: "BPM to use",
		},
		cli.IntFlag{
			Name:  "tick",
			Value: 500,
			Usage: "tick frequency in hertz",
		},
		cli.IntFlag{
			Name:  "hp",
			Value: 65,
			Usage: "high pass note threshold to use for leraning",
		},
		cli.IntFlag{
			Name:  "waits",
			Value: 2,
			Usage: "beats of silence before AI jumps in",
		},
		cli.IntFlag{
			Name:  "quantize",
			Value: 64,
			Usage: "1/quantize is shortest possible note",
		},
		cli.StringFlag{
			Name:  "file,f",
			Value: "music_history.json",
			Usage: "file save/load to when pressing bottom C",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "debug mode",
		},
		cli.BoolFlag{
			Name:  "manual",
			Usage: "AI is activated manually",
		},
		cli.IntFlag{
			Name:  "link",
			Value: 3,
			Usage: "AI LinkLength",
		},
		cli.BoolFlag{
			Name:  "jazzy",
			Usage: "AI Jazziness",
		},
		cli.BoolFlag{
			Name:  "stacatto",
			Usage: "AI Stacattoness",
		},
		cli.BoolFlag{
			Name:  "chords",
			Usage: "AI Allow chords",
		},
		cli.BoolFlag{
			Name:  "follow",
			Usage: "AI velocities follow the host",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		fmt.Println(`
		
		______ _____                   ___  _____ 
		| ___ \_   _|                 / _ \|_   _|
		| |_/ / | |  __ _ _ __   ___ / /_\ \ | |  
		|  __/  | | / _` + "`" + ` | '_ \ / _ \|  _  | | |  
		| |    _| || (_| | | | | (_) | | | |_| |_ 
		\_|    \___/\__,_|_| |_|\___/\_| |_/\___/ 
																							
																							
		 _______________________________________  
		|  | | | |  |  | | | | | |  |  | | | |  | 
		|  | | | |  |  | | | | | |  |  | | | |  | 
		|  | | | |  |  | | | | | |  |  | | | |  | 
		|  |_| |_|  |  |_| |_| |_|  |  |_| |_|  | 
		|   |   |   |   |   |   |   |   |   |   | 
		|   |   |   |   |   |   |   |   |   |   | 
		|___|___|___|___|___|___|___|___|___|___| 

	 Lets play some music!
											`)
		p, err := player.New(c.GlobalInt("bpm"), c.GlobalInt("tick"), c.GlobalBool("debug"))
		if err != nil {
			return
		}
		p.HighPassFilter = c.GlobalInt("hp")
		p.AI = ai2.New(p.TicksPerBeat)
		p.AI.HighPassFilter = c.GlobalInt("hp")
		p.AI.LinkLength = c.GlobalInt("link")
		p.AI.Jazzy = c.GlobalBool("jazzy")
		p.AI.Stacatto = c.GlobalBool("stacatto")
		p.AI.DisallowChords = !c.GlobalBool("chords")
		p.ManualAI = c.GlobalBool("manual")
		p.UseHostVelocity = c.GlobalBool("follow")
		p.Start()
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Print(err)
	}
}
