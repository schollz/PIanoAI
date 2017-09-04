package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/rpiai-piano/player"
	"github.com/urfave/cli"
)

var version string

func main() {

	app := cli.NewApp()
	app.Version = version
	app.Compiled = time.Now()
	app.Name = "rpiai-piano"
	app.Usage = ""
	app.UsageText = `		`
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
		cli.StringFlag{
			Name:  "file,f",
			Value: "music_history.json",
			Usage: "file save/load to when pressing bottom C",
		},
		cli.IntFlag{
			Name:  "link",
			Value: 3,
			Usage: "AI2 LinkLength",
		},
		cli.BoolFlag{
			Name:  "jazzy",
			Usage: "AI2 Jazziness",
		},
		cli.BoolFlag{
			Name:  "stacatto",
			Usage: "AI2 Stacattoness",
		},
		cli.BoolFlag{
			Name:  "chords",
			Usage: "AI2 Allow chords",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		fmt.Println(`
		
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
		p, err := player.New()
		if err != nil {
			return
		}
		p.Start()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Print(err)
	}
}
