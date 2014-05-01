package jlsampler

// #cgo pkg-config: alsa
// #include <alsa/asoundlib.h>
import "C"

import (
	"encoding/binary"
	"fmt"
	"os/exec"
)

func MidiClient(junk interface{}, MsgIn, MsgOut chan *MidiMsg) {
	var handle *C.snd_seq_t
	
	// Open the midi device.
	streams := C.int(C.SND_SEQ_OPEN_INPUT)
	status := int(C.snd_seq_open(&handle, C.CString("default"), streams, 0))
	if status < 0 {
		return
	}
	
	clientNum := int(C.snd_seq_client_id(handle))

	// Give the client a name.
	status = int(C.snd_seq_set_client_name(handle, C.CString("jlsampler")))
	if status < 0 {
		return
	}

	// Create a port.
	name := C.CString("midi_in")
	caps := C.uint(C.SND_SEQ_PORT_CAP_WRITE | C.SND_SEQ_PORT_CAP_SUBS_WRITE)
	type_ := C.uint(C.SND_SEQ_PORT_TYPE_MIDI_GM)

	portNum := int(C.snd_seq_create_simple_port(handle, name, caps, type_))
	if portNum < 0 {
		return
	}

	// Use aconnect to connect midi input. 
	if len(config.MidiIn) != 0 {
		go func() {
			midiPort := fmt.Sprintf("%d:%d", clientNum, portNum)
			cmd := "aconnect " + config.MidiIn + " " + midiPort
			err := exec.Command("sh", "-c", cmd).Run()
			if err != nil {
				Println("Failed aconnect:", cmd, err)
			}
		}()
	}
	
	// Fill message buffer.
	for i := 0; i < config.MidiBufSize; i++ {
		MsgIn <- new(MidiMsg)
	}

	// See seq_event.h for details.
	var ev *C.snd_seq_event_t
	var msg *MidiMsg

	// Process messages.
	for msg = range MsgIn {
		status = int(C.snd_seq_event_input(handle, &ev))
		switch ev._type {

		case C.SND_SEQ_EVENT_NOTEON:
			msg.Type = MsgNoteOn
			msg.SetNote(int8(ev.data[1]) + controls.Transpose)
			msg.Value = float64(ev.data[2]) / 127.0

			// Some controllers send a NoteOn message with zero velocity
			// when a key is released.
			if msg.Value == 0 {
				msg.Type = MsgNoteOff
			}
			MsgOut <- msg

		case C.SND_SEQ_EVENT_NOTEOFF:
			msg.Type = MsgNoteOff
			msg.SetNote(int8(ev.data[1]) + controls.Transpose)
			MsgOut <- msg

		case C.SND_SEQ_EVENT_CONTROLLER:
			msg.Type = MsgControl
			msg.Control = int32(binary.LittleEndian.Uint32(ev.data[4:8]))
			msg.Value = float64(
				int32(binary.LittleEndian.Uint32(ev.data[8:12]))) / 127.0
			MsgOut <- msg

		case C.SND_SEQ_EVENT_PITCHBEND:
			msg.Type = MsgPitchBend
			msg.Value = float64(
				int32(binary.LittleEndian.Uint32(ev.data[8:12]))) / 8192.0
			MsgOut <- msg
		default:
			MsgIn <- msg // Recycle
		}
	}
}
