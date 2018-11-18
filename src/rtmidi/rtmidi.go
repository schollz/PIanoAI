/*
MIT License

Copyright (c) 2017 NaiveSound

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package rtmidi

/*
#cgo CXXFLAGS: -g
#cgo LDFLAGS: -g

#cgo linux CXXFLAGS: -D__LINUX_ALSA__
#cgo linux LDFLAGS: -lasound -pthread
#cgo windows CXXFLAGS: -D__WINDOWS_MM__
#cgo windows LDFLAGS: -luuid -lksuser -lwinmm -lole32
#cgo darwin CXXFLAGS: -D__MACOSX_CORE__
#cgo darwin LDFLAGS: -framework CoreServices -framework CoreAudio -framework CoreMIDI -framework CoreFoundation

#include <stdlib.h>
#include "rtmidi_c.h"

extern void goMIDIInCallback(double ts, unsigned char *msg, size_t msgsz, void *arg);

static inline void midiInCallback(double ts, const unsigned char *msg, size_t msgsz, void *arg) {
	goMIDIInCallback(ts, (unsigned char*) msg, msgsz, arg);
}

static inline void cgoSetCallback(RtMidiPtr in, void *arg) {
	rtmidi_in_set_callback(in, midiInCallback, arg);
}
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

type API C.enum_RtMidiApi

const (
	APIUnspecified API = C.RT_MIDI_API_UNSPECIFIED
	APIMacOSXCore      = C.RT_MIDI_API_MACOSX_CORE
	APILinuxALSA       = C.RT_MIDI_API_LINUX_ALSA
	APIUnixJack        = C.RT_MIDI_API_UNIX_JACK
	APIWindowsMM       = C.RT_MIDI_API_WINDOWS_MM
	APIDummy           = C.RT_MIDI_API_RTMIDI_DUMMY
)

func (api API) String() string {
	switch api {
	case APIUnspecified:
		return "unspecified"
	case APILinuxALSA:
		return "alsa"
	case APIUnixJack:
		return "jack"
	case APIMacOSXCore:
		return "coreaudio"
	case APIWindowsMM:
		return "winmm"
	case APIDummy:
		return "dummy"
	}
	return "?"
}

func CompiledAPI() (apis []API) {
	n := C.rtmidi_get_compiled_api(nil)
	capis := make([]C.enum_RtMidiApi, n, n)
	C.rtmidi_get_compiled_api(&capis[0])
	for _, capi := range capis {
		apis = append(apis, API(capi))
	}
	return apis
}

type MIDI interface {
	OpenPort(port int, name string) error
	OpenVirtualPort(name string) error
	Close() error
	PortCount() (int, error)
	PortName(port int) (string, error)
}

type MIDIIn interface {
	MIDI
	API() (API, error)
	IgnoreTypes(midiSysex bool, midiTime bool, midiSense bool) error
	SetCallback(func(MIDIIn, []byte, float64)) error
	CancelCallback() error
	Message() ([]byte, float64, error)
}

type MIDIOut interface {
	MIDI
	API() (API, error)
	SendMessage([]byte) error
}

type midi struct {
	midi C.RtMidiPtr
}

func (m *midi) OpenPort(port int, name string) error {
	p := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	C.rtmidi_open_port(m.midi, C.uint(port), p)
	if !m.midi.ok {
		return errors.New(C.GoString(m.midi.msg))
	}
	return nil
}

func (m *midi) OpenVirtualPort(name string) error {
	p := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	C.rtmidi_open_virtual_port(m.midi, p)
	if !m.midi.ok {
		return errors.New(C.GoString(m.midi.msg))
	}
	return nil
}

func (m *midi) PortName(port int) (string, error) {
	p := C.rtmidi_get_port_name(m.midi, C.uint(port))
	if !m.midi.ok {
		return "", errors.New(C.GoString(m.midi.msg))
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p), nil
}

func (m *midi) PortCount() (int, error) {
	n := C.rtmidi_get_port_count(m.midi)
	if !m.midi.ok {
		return 0, errors.New(C.GoString(m.midi.msg))
	}
	return int(n), nil
}

func (m *midi) Close() error {
	C.rtmidi_close_port(C.RtMidiPtr(m.midi))
	if !m.midi.ok {
		return errors.New(C.GoString(m.midi.msg))
	}
	return nil
}

type midiIn struct {
	midi
	in C.RtMidiInPtr
	cb func(MIDIIn, []byte, float64)
}

type midiOut struct {
	midi
	out C.RtMidiOutPtr
}

func NewMIDIInDefault() (MIDIIn, error) {
	in := C.rtmidi_in_create_default()
	if !in.ok {
		defer C.rtmidi_in_free(in)
		return nil, errors.New(C.GoString(in.msg))
	}
	return &midiIn{in: in, midi: midi{midi: C.RtMidiPtr(in)}}, nil
}

func NewMIDIIn(api API, name string, queueSize int) (MIDIIn, error) {
	p := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	in := C.rtmidi_in_create(C.enum_RtMidiApi(api), p, C.uint(queueSize))
	if !in.ok {
		defer C.rtmidi_in_free(in)
		return nil, errors.New(C.GoString(in.msg))
	}
	return &midiIn{in: in, midi: midi{midi: C.RtMidiPtr(in)}}, nil
}

func (m *midiIn) API() (API, error) {
	api := C.rtmidi_in_get_current_api(m.in)
	if !m.in.ok {
		return APIUnspecified, errors.New(C.GoString(m.in.msg))
	}
	return API(api), nil
}

func (m *midiIn) Close() error {
	unregisterMIDIIn(m)
	if err := m.midi.Close(); err != nil {
		return err
	}
	C.rtmidi_in_free(m.in)
	return nil
}

func (m *midiIn) IgnoreTypes(midiSysex bool, midiTime bool, midiSense bool) error {
	C.rtmidi_in_ignore_types(m.in, C._Bool(midiSysex), C._Bool(midiTime), C._Bool(midiSense))
	if !m.in.ok {
		return errors.New(C.GoString(m.in.msg))
	}
	return nil
}

var (
	mu     sync.Mutex
	inputs = map[int]*midiIn{}
)

func registerMIDIIn(m *midiIn) int {
	mu.Lock()
	defer mu.Unlock()
	for i := 0; ; i++ {
		if _, ok := inputs[i]; !ok {
			inputs[i] = m
			return i
		}
	}
}

func unregisterMIDIIn(m *midiIn) {
	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < len(inputs); i++ {
		if inputs[i] == m {
			delete(inputs, i)
			return
		}
	}
}

func findMIDIIn(k int) *midiIn {
	mu.Lock()
	defer mu.Unlock()
	return inputs[k]
}

//export goMIDIInCallback
func goMIDIInCallback(ts C.double, msg *C.uchar, msgsz C.size_t, arg unsafe.Pointer) {
	k := int(uintptr(arg))
	m := findMIDIIn(k)
	m.cb(m, C.GoBytes(unsafe.Pointer(msg), C.int(msgsz)), float64(ts))
}

func (m *midiIn) SetCallback(cb func(MIDIIn, []byte, float64)) error {
	k := registerMIDIIn(m)
	m.cb = cb
	C.cgoSetCallback(m.in, unsafe.Pointer(uintptr(k)))
	if !m.in.ok {
		return errors.New(C.GoString(m.in.msg))
	}
	return nil
}

func (m *midiIn) CancelCallback() error {
	unregisterMIDIIn(m)
	C.rtmidi_in_cancel_callback(m.in)
	if !m.in.ok {
		return errors.New(C.GoString(m.in.msg))
	}
	return nil
}

func (m *midiIn) Message() ([]byte, float64, error) {
	msg := make([]C.uchar, 64*1024, 64*1024)
	sz := C.size_t(len(msg))
	r := C.rtmidi_in_get_message(m.in, &msg[0], &sz)
	if !m.in.ok {
		return nil, 0, errors.New(C.GoString(m.in.msg))
	}
	b := make([]byte, int(sz), int(sz))
	for i, c := range msg[:sz] {
		b[i] = byte(c)
	}
	return b, float64(r), nil
}

func NewMIDIOutDefault() (MIDIOut, error) {
	out := C.rtmidi_out_create_default()
	if !out.ok {
		defer C.rtmidi_out_free(out)
		return nil, errors.New(C.GoString(out.msg))
	}
	return &midiOut{out: out, midi: midi{midi: C.RtMidiPtr(out)}}, nil
}

func NewMIDIOut(api API, name string) (MIDIOut, error) {
	p := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	out := C.rtmidi_out_create(C.enum_RtMidiApi(api), p)
	if !out.ok {
		defer C.rtmidi_out_free(out)
		return nil, errors.New(C.GoString(out.msg))
	}
	return &midiOut{out: out, midi: midi{midi: C.RtMidiPtr(out)}}, nil
}

func (m *midiOut) API() (API, error) {
	api := C.rtmidi_out_get_current_api(m.out)
	if !m.out.ok {
		return APIUnspecified, errors.New(C.GoString(m.out.msg))
	}
	return API(api), nil
}

func (m *midiOut) Close() error {
	if err := m.midi.Close(); err != nil {
		return err
	}
	C.rtmidi_out_free(m.out)
	return nil
}

func (m *midiOut) SendMessage(b []byte) error {
	p := C.CBytes(b)
	defer C.free(unsafe.Pointer(p))
	C.rtmidi_out_send_message(m.out, (*C.uchar)(p), C.int(len(b)))
	if !m.out.ok {
		return errors.New(C.GoString(m.out.msg))
	}
	return nil
}
