<img align="right" src="https://rpiai.com/content/images/2017/09/gopher-1.svg" width="180" />

# PIanoAI (Piano AI on a Raspberry Pi)

This is code for providing an augmented piano playing experience. When run, this code will provide computer accompaniment that learns in real-time from the human host pianist. When the host pianist stops playing for a given amount of time, the computer AI will then improvise in the space using the style learned from the host.

# See it

Here's an example of me teaching for ~30 seconds and then jamming with the AI:

[![Playing](http://i.imgur.com/F0piGEz.png)](https://www.youtube.com/watch?v=bvMW71BJofc)

Here are some longer clips of me playing with the AI: [clip #2](https://www.youtube.com/watch?v=vF0uQax56a4) [clip #3](https://www.youtube.com/watch?v=yYuBqUxZtp0)

# Read about it

Read more about it at [rpiai.com/piano](https://rpiai.com/piano/).

# Try it

1. Get a MIDI-enabled keyboard and two-way MIDI adapter
2. Get a Raspberry Pi (however, a Windows / Linux / OS X computer should also work) and connect it to the MIDI keyboard.
3. Build latest version of `libportmidi` (OS X: `brew install portmidi`).

```
sudo apt-get install cmake-curses-gui libasound2-dev
git clone https://github.com/aoeu/portmidi.git
cd portmidi
ccmake .  # press in sequence: c, e, c, e, g
make
sudo make install
```

4. [Download the latest release](https://github.com/schollz/pianoai/releases/latest) OR [install Go](https://golang.org/dl/) and `go get` it:

```
go get -v github.com/schollz/pianoai
```

5. Add `export LD_LIBRARY_PATH=/usr/local/lib` to your `.bashrc`. (Unnessecary if you did not build `portmidi`). Reload bash `source ~/.bashrc` if this is the first time.
6. Play!

```
$ pianoai
      ______ _____                   ___  _____
      | ___ \_   _|                 / _ \|_   _|
      | |_/ / | |  __ _ _ __   ___ / /_\ \ | |
      |  __/  | | / _` | '_ \ / _ \|  _  | | |
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
```

# Roadmap

## Must haves

- [ ] [External script that will start/stop piano based on plugging in Midi](https://raspberrypi.stackexchange.com/questions/19600/is-there-a-way-to-automatically-activate-a-script-when-a-usb-device-connects?newreg=270fe49c413340daa171e1dfdbf96de9)
- [ ] Save sessions as MIDI

## Want haves

- [x] ~~Player changes velocity according to the host~~
- [ ] Integrating more AI routines

## Similar projects

- [Google A.I. Duet](https://github.com/googlecreativelab/aiexperiments-ai-duet)
- [DeepJazz](https://github.com/jisungk/deepjazz)

# Acknowledgements

Thanks to [@egonelbre](https://github.com/egonelbre) for the Gopher graphic.
Thanks to [@rakyll](https://github.com/rakyll) for porting `libportmidi` to Go.

# License

`portmidi` is Licensed under Apache License, Version 2.0.

# Get in touch
 
Like this? Need help? Tweet at me [@yakczar](https://twitter.com/intent/tweet?text=@yakczar%20).
