package goduck

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type KeyProps map[string]int

type Encoder struct {
	scriptFile   string
	outputFile   string
	layoutFile   string
	kprop        KeyProps
	lprop        map[string][]string
	byteArray    bytes.Buffer
	defaultDelay int
}

func main() {
	scriptFile := flag.String("i", "", "DuckyScript file")
	outputFile := flag.String("o", "inject.bin", "Output filename")
	layout := flag.String("l", "us", "Keyboard layout file (us/fr/ca/etc. default: us)")
	flag.Parse()

	if *scriptFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	encoder := NewEncoder(*scriptFile, *outputFile, "resources/"+*layout+".yml")
	if err := encoder.Encode(); err != nil {
		log.Fatal(err)
	}
}

func NewEncoder(scriptFile, outputFile, layoutFile string) *Encoder {
	return &Encoder{
		scriptFile: scriptFile,
		outputFile: outputFile,
		layoutFile: layoutFile,
		kprop:      make(KeyProps),
		lprop:      make(map[string][]string),
	}
}

func (e *Encoder) Encode() error {
	if err := e.loadFiles(); err != nil {
		return err
	}
	return e.encodeToFile()
}

func (e *Encoder) loadFiles() error {
	kpropFile, err := ioutil.ReadFile("resources/default.yml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(kpropFile, &e.kprop); err != nil {
		return err
	}

	lpropFile, err := ioutil.ReadFile(e.layoutFile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(lpropFile, &e.lprop); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) encodeToFile() error {
	file, err := os.Open(e.scriptFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastInst []string
	repeat := false
	loop := 0

	for scanner.Scan() {
		inst := scanner.Text()
		inst = strings.TrimSpace(inst)
		if inst == "" || strings.HasPrefix(inst, "//") {
			continue
		}

		if repeat {
			inst = strings.Join(lastInst, " ")
		} else {
			lastInst = strings.Fields(inst)
			loop = 1
			repeat = false
		}

		fields := strings.Fields(inst)
		if len(fields) == 0 {
			continue
		}
		cmd := fields[0]

		if cmd == "REM" {
			continue
		} else if cmd == "REPEAT" {
			repeatCount, err := strconv.Atoi(fields[1])
			if err != nil {
				return err
			}
			loop = repeatCount
			repeat = true
		} else if cmd == "DEFAULT_DELAY" || cmd == "DEFAULTDELAY" {
			delay, err := strconv.Atoi(fields[1])
			if err != nil {
				return err
			}
			e.defaultDelay = delay
		} else {
			for i := 0; i < loop; i++ {
				if err := e.processInstruction(fields); err != nil {
					return err
				}
				if e.defaultDelay > 0 && cmd != "DELAY" {
					if err := e.addDelay(e.defaultDelay); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return ioutil.WriteFile(e.outputFile, e.byteArray.Bytes(), 0644)
}

func (e *Encoder) processInstruction(fields []string) error {
	cmd := fields[0]
	switch cmd {
	case "DELAY":
		delay, err := strconv.Atoi(fields[1])
		if err != nil {
			return err
		}
		return e.addDelay(delay)
	case "STRING":
		for _, char := range fields[1] {
			bytes, err := e.charToBytes(string(char))
			if err != nil {
				return err
			}
			e.byteArray.Write(bytes)
		}
	default:
		code, err := e.instToByte(cmd)
		if err != nil {
			return err
		}
		e.byteArray.WriteByte(byte(code))
		e.byteArray.WriteByte(0x00)
	}
	return nil
}

func (e *Encoder) addDelay(ms int) error {
	for ms > 0 {
		e.byteArray.WriteByte(0x00)
		if ms > 255 {
			e.byteArray.WriteByte(0xFF)
			ms -= 255
		} else {
			e.byteArray.WriteByte(byte(ms))
			ms = 0
		}
	}
	return nil
}

func (e *Encoder) charToBytes(char string) ([]byte, error) {
	code := e.charToCode(char)
	return e.codeToBytes(code)
}

func (e *Encoder) charToCode(char string) string {
	r := []rune(char)[0]
	if r < 128 {
		return fmt.Sprintf("ASCII_%X", r)
	} else if r < 256 {
		return fmt.Sprintf("ISO_8859_1_%X", r)
	} else {
		return fmt.Sprintf("UNICODE_%X", r)
	}
}

func (e *Encoder) codeToBytes(code string) ([]byte, error) {
	keys, ok := e.lprop[code]
	if !ok {
		return nil, fmt.Errorf("character not found: %s", code)
	}

	var bytes []byte
	for _, key := range keys {
		if val, ok := e.kprop[key]; ok {
			bytes = append(bytes, byte(val))
		} else {
			return nil, fmt.Errorf("key not found: %s", key)
		}
	}
	return bytes, nil
}

func (e *Encoder) instToByte(instruction string) (int, error) {
	key := "KEY_" + strings.ToUpper(instruction)
	if val, ok := e.kprop[key]; ok {
		return val, nil
	}
	switch instruction {
	case "ESCAPE":
		return e.instToByte("ESC")
	case "DEL":
		return e.instToByte("DELETE")
	case "BREAK":
		return e.instToByte("PAUSE")
	case "CONTROL":
		return e.instToByte("CTRL")
	case "DOWNARROW":
		return e.instToByte("DOWN")
	case "UPARROW":
		return e.instToByte("UP")
	case "LEFTARROW":
		return e.instToByte("LEFT")
	case "RIGHTARROW":
		return e.instToByte("RIGHT")
	case "MENU":
		return e.instToByte("APP")
	case "WINDOWS":
		return e.instToByte("GUI")
	case "PLAY", "PAUSE":
		return e.instToByte("MEDIA_PLAY_PAUSE")
	case "STOP":
		return e.instToByte("MEDIA_STOP")
	case "MUTE":
		return e.instToByte("MEDIA_MUTE")
	case "VOLUMEUP":
		return e.instToByte("MEDIA_VOLUME_INC")
	case "VOLUMEDOWN":
		return e.instToByte("MEDIA_VOLUME_DEC")
	case "SCROLLLOCK":
		return e.instToByte("SCROLL_LOCK")
	case "NUMLOCK":
		return e.instToByte("NUM_LOCK")
	case "CAPSLOCK":
		return e.instToByte("CAPS_LOCK")
	}
	bytes, err := e.charToBytes(string(instruction[0]))
	if err != nil {
		return 0, err
	}
	return int(bytes[0]), nil
}
