package main

import (
	"flag"
	"fmt"
	"github.com/NubeDev/modbus"
	"net"
	"os"
	"os/signal"
	"time"
)

var _ = time.Second

var fillData = flag.String("d", "am3", "data to start with, am3 starts memory "+
	"with bools as address (mod 3) == 0, and registers as address * 3 (mod uint16)")

var writeSizeLimit = flag.Int("wsl", modbus.MaxRTUSize, "client only, the max size in bytes of a write to server to send")
var readSizeLimit = flag.Int("rsl", modbus.MaxRTUSize, "client only, the max size in bytes of a read from server to request")

var verbose = flag.Bool("v", false, "prints debugging information")

func handlerGenerator(name string) modbus.ProtocolHandler {
	return &modbus.SimpleHandler{
		ReadHoldingRegisters: func(address, quantity uint16) ([]uint16, error) {
			fmt.Printf("%v ReadHoldingRegisters from %v, quantity %v\n",
				name, address, quantity)
			r := make([]uint16, quantity)
			// application code that fills in r here
			return r, nil
		},
		WriteHoldingRegisters: func(address uint16, values []uint16) error {
			fmt.Printf("%v WriteHoldingRegisters from %v, quantity %v\n",
				name, address, len(values))
			// application code here
			return nil
		},
		OnErrorImp: func(req modbus.PDU, errRep modbus.PDU) {
			fmt.Printf("%v received error:%x in request:%x", name, errRep, req)
		},
	}
}

func main() {
	flag.Parse()
	if *verbose {
		modbus.SetDebugOut(os.Stdout)
	}

	// TCP address of the host
	host := "0.0.0.0:10502"

	// Default server id
	//id := byte(1)

	// Open server tcp listener:
	listener, err := net.Listen("tcp", host)
	if err != nil {
		fmt.Println(err)
		return
	}

	com := modbus.NewTCPServer(listener)

	defer func() {
		com.Close()
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		fmt.Println("close server port")
		com.Close()
		os.Exit(0)
	}()

	if *fillData == "am3" {
		fillAm3()
	}
	var device modbus.Server
	device = com
	h := modbus.SimpleHandler{
		ReadDiscreteInputs: func(address, quantity uint16) ([]bool, error) {
			fmt.Printf("ReadDiscreteInputs from %v, quantity %v\n", address, quantity)
			return discretes[address : address+quantity], nil
		},
		WriteDiscreteInputs: func(address uint16, values []bool) error {
			fmt.Printf("WriteDiscreteInputs from %v, quantity %v\n", address, len(values))
			for i, v := range values {
				discretes[address+uint16(i)] = v
			}
			return nil
		},

		ReadCoils: func(address, quantity uint16) ([]bool, error) {
			fmt.Printf("ReadCoils from %v, quantity %v\n", address, quantity)
			return coils[address : address+quantity], nil
		},
		WriteCoils: func(address uint16, values []bool) error {
			fmt.Printf("WriteCoils from %v, quantity %v\n", address, len(values))
			for i, v := range values {
				coils[address+uint16(i)] = v
				fmt.Println(i, v)
			}
			return nil
		},

		ReadInputRegisters: func(address, quantity uint16) ([]uint16, error) {
			fmt.Printf("ReadInputRegisters from %v, quantity %v\n", address, quantity)
			return inputRegisters[address : address+quantity], nil
		},
		WriteInputRegisters: func(address uint16, values []uint16) error {
			fmt.Printf("WriteInputRegisters from %v, quantity %v\n", address, len(values))
			for i, v := range values {
				inputRegisters[address+uint16(i)] = v
			}
			return nil
		},

		ReadHoldingRegisters: func(address, quantity uint16) ([]uint16, error) {
			fmt.Printf("ReadHoldingRegisters from %v, quantity %v\n", address, quantity)
			return holdingRegisters[address : address+quantity], nil
		},
		WriteHoldingRegisters: func(address uint16, values []uint16) error {
			fmt.Printf("WriteHoldingRegisters from %v, quantity %v\n", address, len(values))
			for i, v := range values {
				holdingRegisters[address+uint16(i)] = v
			}
			return nil
		},

		OnErrorImp: func(req modbus.PDU, errRep modbus.PDU) {
			fmt.Printf("error received: %v from req: %v\n", errRep, req)
		},
	}
	err = device.Serve(&h)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %v\n", err)
		os.Exit(1)
	}
}

const size = 0x10000

var discretes [size]bool
var coils [size]bool
var inputRegisters [size]uint16
var holdingRegisters [size]uint16

func fillAm3() {
	for i := range discretes {
		discretes[i] = false
	}
	for i := range coils {
		coils[i] = false
	}
	for i := range inputRegisters {
		inputRegisters[i] = uint16(0)
	}
	for i := range holdingRegisters {
		holdingRegisters[i] = uint16(0)
	}
}
