package jlsampler

import (
	"github.com/johnnylee/jackclient"
)

// ----------------------------------------------------------------------------
type JackPlayer struct {
	name string
	client *jackclient.JackClient
	buf *Sound

	BufIn chan *Sound
	BufOut chan *Sound
}

func NewJackPlayer(name string) *JackPlayer {
	var err error
	jp := new(JackPlayer)
	jp.name = name
	jp.buf = NewSound(0)
	jp.client, err = jackclient.New(jp.name, 0, 2)
	if err != nil {
		panic("Failed to create jack client!")
	}
	return jp
}

func (jp *JackPlayer) Run(junk interface{}, BufIn, BufOut chan *Sound) {
	jp.BufIn = BufIn
	jp.BufOut = BufOut
	jp.client.RegisterCallback(jp.process)
}

func (jp *JackPlayer) process(bufIn, bufOut [][]float32) error {
	L := bufOut[0]
	R := bufOut[1]
	
	// Create new buffer if necessary. 
	jp.buf.L = L
	jp.buf.R = R
	jp.buf.Len = len(L)
	
	// Send and receive buffer. 
	jp.BufOut <- jp.buf
	jp.buf = <-jp.BufIn
	
	return nil 
}
