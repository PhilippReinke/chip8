package emulator

import (
	"fmt"
	"math/rand/v2"
)

const (
	FontStart    = 0x50
	ProgramStart = 0x200

	DisplayWidth  = 64
	DisplayHeight = 32
	DisplaySize   = DisplayWidth * DisplayHeight
)

// Emulator holds memory, registers, display and timer.
type Emulator struct {
	memory [4096]byte
	v      [16]byte // V0 ... VF
	i      uint16   // index register
	pc     uint16   // program counter

	sp    byte       // stack pointer
	stack [16]uint16 // stack

	DrawFlag bool
	Display  [DisplaySize]byte
	keypad   [16]bool

	delayTimer byte
	soundTimer byte
}

// New creates CHIP-8 emulator.
func New() *Emulator {
	return &Emulator{
		pc: ProgramStart,
	}
}

// LoadROM load given ROM in memory.
func (e *Emulator) LoadROM(data []byte) {
	copy(e.memory[ProgramStart:], data)
}

// UpdateTimers delay and sound timer.
func (e *Emulator) UpdateTimers() {
	if e.delayTimer > 0 {
		e.delayTimer--
	}

	if e.soundTimer > 0 {
		if e.soundTimer == 1 {
			fmt.Println("beep")
		}
		e.soundTimer--
	}
}

// EmulateCycle fetches, decodes and executes instruction.
func (e *Emulator) EmulateCycle() {
	opcode := uint16(e.memory[e.pc])<<8 | uint16(e.memory[e.pc+1])
	e.pc += 2 // increase program count

	// NNN: address
	// NN: 8-bit constant
	// N: 4-bit constant
	// X and Y: 4-bit register identifier
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := byte(opcode & 0x00FF)
	nnn := opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // CLS
			for i := range e.Display {
				e.Display[i] = 0
			}
			e.DrawFlag = true
		case 0x00EE: // RET
			e.sp--
			e.pc = e.stack[e.sp]
		default:
			fmt.Printf("Unknown 0x0000 opcode: 0x%X\n", opcode)
		}

	case 0x1000: // JP addr
		e.pc = nnn

	case 0x2000: // CALL addr
		e.stack[e.sp] = e.pc
		e.sp++
		e.pc = nnn

	case 0x3000: // SE Vx, byte
		if e.v[x] == nn {
			e.pc += 2
		}

	case 0x4000: // SNE Vx, byte
		if e.v[x] != nn {
			e.pc += 2
		}

	case 0x5000: // SE Vx, Vy
		if e.v[x] == e.v[y] {
			e.pc += 2
		}

	case 0x6000: // LD Vx, byte
		e.v[x] = nn

	case 0x7000: // ADD Vx, byte
		e.v[x] += nn

	case 0x8000:
		switch opcode & 0x000F {
		case 0x0: // LD Vx, Vy
			e.v[x] = e.v[y]
		case 0x1: // OR Vx, Vy
			e.v[x] |= e.v[y]
		case 0x2: // AND Vx, Vy
			e.v[x] &= e.v[y]
		case 0x3: // XOR Vx, Vy
			e.v[x] ^= e.v[y]
		case 0x4: // ADD Vx, Vy
			sum := uint16(e.v[x]) + uint16(e.v[y])
			e.v[0xF] = 0
			if sum > 255 {
				e.v[0xF] = 1
			}
			e.v[x] = byte(sum)
		case 0x5: // SUB Vx, Vy
			e.v[0xF] = 0
			if e.v[x] > e.v[y] {
				e.v[0xF] = 1
			}
			e.v[x] -= e.v[y]
		case 0x6: // SHR Vx
			e.v[0xF] = e.v[x] & 0x1
			e.v[x] >>= 1
		case 0x7: // SUBN Vx, Vy
			e.v[0xF] = 0
			if e.v[y] > e.v[x] {
				e.v[0xF] = 1
			}
			e.v[x] = e.v[y] - e.v[x]
		case 0xE: // SHL Vx
			e.v[0xF] = (e.v[x] & 0x80) >> 7
			e.v[x] <<= 1
		}

	case 0x9000: // SNE Vx, Vy
		if e.v[x] != e.v[y] {
			e.pc += 2
		}

	case 0xA000: // LD I, addr
		e.i = nnn

	case 0xB000: // JP V0, addr
		e.pc = uint16(e.v[0]) + nnn

	case 0xC000: // RND Vx, byte
		e.v[x] = byte(rand.IntN(256)) & nn

	case 0xD000: // DRW Vx, Vy, nibble
		vx := e.v[x]
		vy := e.v[y]
		height := n
		e.v[0xF] = 0

		for row := uint16(0); row < height; row++ {
			pixel := e.memory[e.i+row]
			for col := uint16(0); col < 8; col++ {
				if (pixel & (0x80 >> col)) != 0 {
					index := ((int(vy)+int(row))%DisplayHeight)*DisplayWidth + (int(vx)+int(col))%DisplayWidth
					if e.Display[index] == 1 {
						e.v[0xF] = 1
					}
					e.Display[index] ^= 1
				}
			}
		}
		e.DrawFlag = true

	case 0xE000:
		switch opcode & 0x00FF {
		case 0x9E: // SKP Vx
			if e.keypad[e.v[x]] {
				e.pc += 2
			}
		case 0xA1: // SKNP Vx
			if !e.keypad[e.v[x]] {
				e.pc += 2
			}
		}

	case 0xF000:
		switch opcode & 0x00FF {
		case 0x07:
			e.v[x] = e.delayTimer
		case 0x0A:
			// Wait for key press
			keyPressed := false
			for i, k := range e.keypad {
				if k {
					e.v[x] = byte(i)
					keyPressed = true
					break
				}
			}
			if !keyPressed {
				e.pc -= 2 // stay on this instruction
			}
		case 0x15:
			e.delayTimer = e.v[x]
		case 0x18:
			e.soundTimer = e.v[x]
		case 0x1E:
			e.i += uint16(e.v[x])
		case 0x29:
			e.i = FontStart + uint16(e.v[x])*5
		case 0x33:
			vx := e.v[x]
			e.memory[e.i] = vx / 100
			e.memory[e.i+1] = (vx / 10) % 10
			e.memory[e.i+2] = vx % 10
		case 0x55:
			for i := uint16(0); i <= x; i++ {
				e.memory[e.i+i] = e.v[i]
			}
		case 0x65:
			for i := uint16(0); i <= x; i++ {
				e.v[i] = e.memory[e.i+i]
			}
		}

	default:
		fmt.Printf("Unknown opcode: 0x%X\n", opcode)
	}
}
