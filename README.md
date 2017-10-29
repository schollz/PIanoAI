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

1. Get a MIDI-enabled keyboard and [two-way MIDI adapter ($18)](https://www.amazon.com/gp/product/B0739M6XZ1/ref=as_li_tl?ie=UTF8&tag=scholl-20&camp=1789&creative=9325&linkCode=as2&creativeASIN=B0739M6XZ1&linkId=a780a9b1af1417680e01e183c137be1d). Or get a [USB-MIDI keyboard ($38)](https://www.amazon.com/gp/product/B00VHKMK64/ref=as_li_tl?ie=UTF8&tag=scholl-20&camp=1789&creative=9325&linkCode=as2&creativeASIN=B00VHKMK64&linkId=51809da99cc2145b572498639b367c9c).
2. Get a [Raspberry Pi](https://www.amazon.com/gp/product/B01C6EQNNK/ref=as_li_tl?ie=UTF8&tag=scholl-20&camp=1789&creative=9325&linkCode=as2&creativeASIN=B01C6EQNNK&linkId=805012388be781415a6be827b50c76ac) (however, a Windows / Linux / OS X computer should also work) and connect it to the MIDI keyboard.
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
```

For some extra jazziness do

```
$ pianoai --jazzy
```

## Options 

### Piano keyboard controls

When you play, you can always trigger learning and improvising by hitting the top B or top C respectively, on the piano keyboard (assuming an 88-key keyboard). If you use `--manual` mode then you can only hear improvisation after triggering. Normally, however, the improvisation will start as soon as it has enough notes and you leave enough space for the improvisation to take place (usually a few beats).

You can save your current data by pressing the bottom A on the piano keyboard and you can play back what *you* played by hitting the bottom Bb on the piano keyboard. Currently there is not a way to save the AI playing (but its in the roadmap, see below).

### Command line options

There are many command-line options for tuning the AI, but feel free to play with the code as well. Current options:

```
   --bpm value             BPM to use (default: 120)
   --tick value            tick frequency in hertz (default: 500)
   --hp value              high pass note threshold to use for leraning (default: 65)
   --waits value           beats of silence before AI jumps in (default: 2)
   --quantize value        1/quantize is shortest possible note (default: 64)
   --file value, -f value  file save/load to when pressing bottom C (default: "music_history.json")
   --debug                 debug mode
   --manual                AI is activated manually
   --link value            AI LinkLength (default: 3)
   --jazzy                 AI Jazziness
   --stacatto              AI Stacattoness
   --chords                AI Allow chords
   --follow                AI velocities follow the host
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
