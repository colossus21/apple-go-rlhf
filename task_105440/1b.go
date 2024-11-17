//go:build 1b
// +build 1b

package main

import (
	"fmt"
	"time"
)

// States of the Lightsaber
const (
	LIGHTSABER_OFF = iota
	LIGHTSABER_ON
	LIGHTSABER_HEATING
)

// Sabertype for choosing colours
const (
	JEDI_LIGHTSABER = iota
	SITH_LIGHTSABER
)

type Lightsaber struct {
	currentState int
	sabertype    int
}

func (l *Lightsaber) initialize() {
	l.currentState = LIGHTSABER_OFF
	l.sabertype = JEDI_LIGHTSABER
}

// Transitions the Lightsaber between states
func (l *Lightsaber) transition(event string) {
	switch l.currentState {
	case LIGHTSABER_OFF:
		switch event {
		case "ON":
			l.currentState = LIGHTSABER_HEATING
		}
	case LIGHTSABER_HEATING:
		switch event {
		case "ON":
			l.currentState = LIGHTSABER_ON
			fmt.Println("The Lightsaber is activated!")
			if l.sabertype == JEDI_LIGHTSABER {
				fmt.Println("Blue Light illuminated!")
			} else {
				fmt.Println("Red Light illuminated!")
			}
		}
	case LIGHTSABER_ON:
		switch event {
		case "OFF":
			l.currentState = LIGHTSABER_OFF
			fmt.Println("The Lightsaber is deactivated.")
		}
	}
}

func (l *Lightsaber) changesaberType(sabertype int) {
	if sabertype == JEDI_LIGHTSABER || sabertype == SITH_LIGHTSABER {
		l.sabertype = sabertype
		fmt.Println("Saber changed type!")
	} else {
		fmt.Println("Not a valid lightsaber type!")
	}
}

func main() {
	lightsaber := &Lightsaber{}
	lightsaber.initialize()

	time.Sleep(500 * time.Millisecond) //Simulate press of button
	lightsaber.transition("ON")

	time.Sleep(2000 * time.Millisecond)
	lightsaber.transition("ON")

	time.Sleep(2000 * time.Millisecond)
	lightsaber.transition("OFF")

	time.Sleep(500 * time.Millisecond)
	lightsaber.changesaberType(SITH_LIGHTSABER)

	time.Sleep(500 * time.Millisecond)
	lightsaber.transition("ON")
}
