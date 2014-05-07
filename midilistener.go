package jlsampler

// #cgo pkg-config: alsa
// #include <alsa/asoundlib.h>
import "C"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os/exec"
)

// ----------------------------------------------------------------------------
type MidiListener struct {
	handle *C.snd_seq_t // Handle for midi sequencer.
}

/* NewMidiListener
 * name     : The name of the device.
 * midiPort : An incoming midi port to connect to.
 */
func NewMidiListener(name, midiPort string) (*MidiListener, error) {
	ml := new(MidiListener)

	// Open the midi device.
	openIn := C.int(C.SND_SEQ_OPEN_INPUT)
	status := int(C.snd_seq_open(&ml.handle, C.CString("default"), openIn, 0))
	if status < 0 {
		return nil, errors.New("Failed to open midi device.")
	}

	clientNum := int(C.snd_seq_client_id(ml.handle))

	// Give the client a name.
	status = int(C.snd_seq_set_client_name(ml.handle, C.CString(name)))
	if status < 0 {
		return nil, errors.New("Failed to set client name.")
	}

	// Create a port.
	caps := C.uint(C.SND_SEQ_PORT_CAP_WRITE | C.SND_SEQ_PORT_CAP_SUBS_WRITE)
	type_ := C.uint(C.SND_SEQ_PORT_TYPE_MIDI_GM)

	portNum := int(
		C.snd_seq_create_simple_port(ml.handle, C.CString(name), caps, type_))
	if portNum < 0 {
		return nil, errors.New("Failed to create port.")
	}

	// Call aconnect to connect midi port.
	if len(midiPort) != 0 {
		go func() {
			midiPort := fmt.Sprintf("%d:%d", clientNum, portNum)
			cmd := "aconnect " + config.MidiIn + " " + midiPort
			err := exec.Command("sh", "-c", cmd).Run()
			if err != nil {
				Println("Failed in call to aconnect:", cmd, err)
			}
		}()
	}

	return ml, nil
}

func noteAndValue(ev *C.snd_seq_event_t) (int8, float64) {
	note := int8(ev.data[1]) + controls.Transpose
	value := float64(ev.data[2]) / 127.0
	return note, value
}

/* Run
 * Read incoming midi events and send them to the sampler.
 */
func (ml *MidiListener) Run() {
	var ev *C.snd_seq_event_t
	var status int
	var note int8
	var value float64

	for {
		status = int(C.snd_seq_event_input(ml.handle, &ev))
		if status < 0 {
			Println("Error reading midi event. Ignoring.")
			continue
		}

		switch ev._type {

		case C.SND_SEQ_EVENT_NOTEON:
			note, value = noteAndValue(ev)
			if ev.data[2] != 0 {
				sampler.NoteOnEvent(note, value)
			} else {
				sampler.NoteOffEvent(note, value)
			}

		case C.SND_SEQ_EVENT_NOTEOFF:
			note, value = noteAndValue(ev)
			sampler.NoteOffEvent(note, value)

		case C.SND_SEQ_EVENT_CONTROLLER:
			sampler.ControllerEvent(
				int32(binary.LittleEndian.Uint32(ev.data[4:8])),
				float64(binary.LittleEndian.Uint32(ev.data[8:12]))/127)

		case C.SND_SEQ_EVENT_PITCHBEND:
			sampler.PitchBendEvent(
				float64(binary.LittleEndian.Uint32(ev.data[8:12])) / 8192.0)
		}
	}
}