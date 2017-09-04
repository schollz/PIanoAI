package main

import (
	"fmt"

	"github.com/schollz/rpiai-piano/player"
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

func main() {
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
	var err error
	p, err := player.New(240)
	if err != nil {
		panic(err)
	}
	p.Start()
}
