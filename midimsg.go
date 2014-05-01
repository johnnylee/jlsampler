package jlsampler

const (
	MsgNoteOn    = iota
	MsgNoteOff   = iota
	MsgPitchBend = iota
	MsgControl   = iota
)

type MidiMsg struct {
	Type    int8  // Constant: NoteOn, NoteOff, PitchBend, Control.
	Note    int8  // If NoteOn or NoteOff, this is the note (0-127).
	Control int32 // If Type is Control, then this is the control (1-?).

	// The note velocity or control value.
	// Rnages:
	// NoteOne (velocity): [0-1]
	// PitchBend: [-1, 1]
	// Control: [0, 1]
	Value float64
}

func (m *MidiMsg) SetNote(note int8) {
	if note > 127 {
		note = 127
	} else if note < 1 {
		note = 1
	}
	m.Note = note
}