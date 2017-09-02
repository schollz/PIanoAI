<img align="right" src="https://rpiai.com/content/images/2017/09/gopher-1.svg" width="180" />

# Raspberry Pi Piano Player AI

This is code for providing an augmented piano playing experience. When run, this code will provide computer accompaniment that learns in real-time from the human host pianist. Then the host pianist stops playing for a given amount of time, the computer AI will then improvise in the space using the style learned from the host.


# Try it

1. Get a MIDI-enabled keyboard and two-way MIDI adapter
2. Get a Raspberry Pi (however, a Windows / Linux / OS X computer should also work) and connect it to the MIDI keyboard.
3. Build latest version of `libportmidi` (if your using Mac just do `brew install portmidi`, if Windows just [Download it](https://sourceforge.net/projects/portmedia/files/portmidi/217/pmdefaults-setup-w32-217.zip/download)))

```
sudo apt-get install cmake-curses-gui libasound2-dev
git clone https://github.com/aoeu/portmidi.git
cd portmidi
ccmake .  # press in sequence: c, e, c, e, g
make
sudo make install
```

4. [Install Go](https://golang.org/dl/).
5. Install `rpiai-piano`:

```
go get -v github.com/schollz/rpiai-piano
```

6. Add `export LD_LIBRARY_PATH=/usr/local/lib` to your `.bashrc`. (Unnessecary if you did not build `portmidi`). Reload bash `source ~/.bashrc` if this is the first time.
7. Play!

```
rpiai-piano
```

## TODO

- [ ] [External script that will start/stop piano based on plugging in Midi](https://raspberrypi.stackexchange.com/questions/19600/is-there-a-way-to-automatically-activate-a-script-when-a-usb-device-connects?newreg=270fe49c413340daa171e1dfdbf96de9)
- [ ] Put links to code on Youtube videos (link to tree)
- [ ] If a config file is not present, use default values and create one for next time (and send a message to the user)
- [ ] Add a new button for reseting the system
- [ ] Add a function for shutting down
- [ ] Add command line functions

# Acknowledgements

Thanks to [@egonelbre](https://github.com/egonelbre) for the Gopher graphic.
Thanks to [@rakyll](https://github.com/rakyll) for porting `libportmidi` to Go.

# License

`portmidi` is Licensed under Apache License, Version 2.0.

