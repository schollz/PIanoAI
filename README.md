<img align="right" src="https://rpiai.com/content/images/2017/09/gopher-1.svg" width="180" />

# Raspberry Pi Piano Player AI

This is code for providing an augmented piano playing experience. When run, this code will provide computer accompaniment that learns in real-time from the human host pianist. Then the host pianist stops playing for a given amount of time, the computer AI will then improvise in the space using the style learned from the host.



# Install

## Materials

1. MIDI-enabled keyboard and two-way MIDI adapter
2. A Raspberry Pi (though Windows / Linux / OS X should also work).

## Setup

3. [install Go](https://golang.org/dl/) on the computer you will use
4. Install the `libportmidi` (v. 217) library: 

### Linux

`apt-get install libportmidi-dev`

### OS X

`brew install portmidi`

### Windows

[Download ](https://sourceforge.net/projects/portmedia/files/portmidi/217/pmdefaults-setup-w32-217.zip/download) from [Sourceforge](https://sourceforge.net/projects/portmedia/files/portmidi/217/).



## TODO

- [ ] [External script that will start/stop piano based on plugging in Midi](https://raspberrypi.stackexchange.com/questions/19600/is-there-a-way-to-automatically-activate-a-script-when-a-usb-device-connects?newreg=270fe49c413340daa171e1dfdbf96de9)
- [ ] Allow threshold for using notes for AI. (i.e. AI only generates licks on above middle C, or similar)
- [ ] Implement minimum distance between notes in AI so that more starts/stops are available.
- [ ] Put links to code on Youtube videos (link to tree)

# Acknowledgements

Thanks to [@egonelbre](https://github.com/egonelbre) for the Gopher graphic.
Thanks to [@rakyll](https://github.com/rakyll) for porting `libportmidi` to Go.

# License

`portmidi` is Licensed under Apache License, Version 2.0.

