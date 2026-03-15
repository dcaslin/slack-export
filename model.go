package main

import "time"

type SimpleMsg struct {
	Name string
	Text string
	Ts   time.Time
}
