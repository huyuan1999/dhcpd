package main

import (
	"testing"
)

func TestHandler(t *testing.T) {
	o := &Options{
		LeaseTime:    "1h",
		ServerIP:     "10.1.1.1",
		BootFileName: "pxelinux.0",
		GatewayIP:    "10.1.1.1",
		RangeStartIP: "10.1.1.100",
		RangeEndIP:   "10.1.1.110",
		NetMask:      "255.0.0.0",
		Router:       "10.1.1.2, 10.1.1.3",
		DNS:          "10.1.1.4, 10.1.1.5",
	}
	connectDB()
	db.Create(o)
}
