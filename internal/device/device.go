package device

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type deviceState int

// represents device states
const (
	deviceReady deviceState = iota
	deviceBusy
	deviceError
)

const (
	networkWifi string = "tcp"
)

type RawConnection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

// Represents a bare device with minimum config for lowlevel operations
type RealDevice struct {
	mutex   sync.Mutex
	state   deviceState
	input   string
	outputs []string
	conn    RawConnection
	version string
}

// Represents a wifi device
type WifiDevice struct {
	device *RealDevice
}

func (d *RealDevice) write(data string) (int, error) {
	d.input = ""

	n, err := d.conn.Write([]byte(data + "\r\n"))

	if err == nil {
		d.input = data
	}

	return n, err

}

func (d *RealDevice) read() error {
	var buffer bytes.Buffer

	ticker := time.NewTicker(10 * time.Millisecond)

	for range ticker.C {
		byteSeq := make([]byte, 128)
		n, err := d.conn.Read(byteSeq)
		if err != nil {
			d.outputs = []string{}
			return err
		}

		buffer.Write(byteSeq[:n])
		if string(byteSeq[n-1]) == ">" {
			buffer.Truncate(buffer.Len() - 1)
			ticker.Stop()
			break
		}
	}
	return d.processResult(buffer)
}

func (d *RealDevice) processResult(buf bytes.Buffer) error {
	parts := strings.Split(buf.String(), "\r")

	if parts[0] != d.input {
		return fmt.Errorf("command echo mismatch: %q not suffix of %q",
			d.input, parts[0])
	}

	parts = parts[1:]

	var completeParts []string

	for p := range parts {
		data := strings.Trim(parts[p], "\r")
		if data == "" {
			continue
		}

		completeParts = append(completeParts, data)
	}

	if len(completeParts) < 1 {
		return fmt.Errorf("not data returned by ECU")
	}

	d.outputs = completeParts

	return nil
}

func (d *RealDevice) RunCommand(cmd string) *Result {
	var (
		err     error
		tsTotal time.Time
		tsWrite time.Time
		tsRead  time.Time
	)

	result := Result{
		input:     cmd,
		readTime:  0,
		writeTime: 0,
		totalTime: 0,
	}

	tsTotal = time.Now()
	d.mutex.Lock()
	d.state = deviceBusy

	tsWrite = time.Now()
	_, err = d.write(cmd)
	if err != nil {
		goto out
	}

	result.writeTime = time.Since(tsWrite)

	err = d.read()
	result.readTime = time.Since(tsRead)

	if err != nil {
		goto out
	}

out:
	if err != nil {
		d.state = deviceError
	} else {
		d.state = deviceReady
	}
	d.mutex.Unlock()
	result.err = err
	result.outputs = d.outputs
	result.totalTime = time.Since(tsTotal)

	return &result
}

func (d *RealDevice) Reset() error {

	result := d.RunCommand("ATZ")

	if result.Failed() {
		return result.GetError()
	}

	outputs := result.GetOutput()

	if (strings.HasPrefix(outputs[0], "ELM327")) || (len(outputs) > 1 && strings.HasPrefix(outputs[1], "ELM327")) {
		output := outputs[0]
		if len(outputs) > 1 {
			output += " " + outputs[1]
		}
		err := fmt.Errorf("device failed to identify as ELM327: %s", output)
		d.state = deviceError
		return err
	}

	err := d.setVersion()
	if err != nil {
		d.state = deviceError
		return err
	}
	return nil
}

func (d *RealDevice) setAutoProtocol() error {
	result := d.RunCommand("ATSP0")

	if result.Failed() {
		return result.GetError()
	}

	if result.GetOutput()[0] != "OK" {
		return fmt.Errorf("protocol setup failed. Got response : %q", result.GetOutput()[0])
	}

	return nil

}

func (d *RealDevice) setVersion() error {
	pat := regexp.MustCompile("v[0-9]+.[0-9]+")
	for i := range d.outputs {
		if m := pat.Find([]byte(d.outputs[i])); m != nil {
			d.version = string(m)
			return nil
		}
	}
	return fmt.Errorf("failed to identify device version. Bailing out early")
}

func (d *RealDevice) getVersion() string {
	return d.version
}
