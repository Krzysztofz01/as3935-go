package as3935go

import (
	"fmt"
	"math"
	"sync"
	"time"

	"golang.org/x/exp/io/i2c"
)

type IRQOutputSource uint8

const (
	None IRQOutputSource = 0x00
	TRCO IRQOutputSource = 0x20
	SRCO IRQOutputSource = 0x40
	LCO  IRQOutputSource = 0x80
)

type InterruptType uint8

const (
	NoResults          InterruptType = 0x00
	NoiseLevelTooHigh  InterruptType = 0x01
	DisturberDetected  InterruptType = 0x04
	LightningInterrupt InterruptType = 0x08
)

type TuningCapacitance uint16

const (
	TuningDiv16  TuningCapacitance = 0x0000
	TuningDiv32  TuningCapacitance = 0x000F
	TuningDiv64  TuningCapacitance = 0x0F00
	TuningDiv128 TuningCapacitance = 0x0F0F
)

type AnalogFrontEnd uint8

const (
	Indoor  AnalogFrontEnd = 0x24
	Outdoor AnalogFrontEnd = 0x1C
)

const delayDuration = time.Duration(5) * time.Millisecond

type Module interface {
	// Open the communication with the module over i2c.
	Open() error

	// Close the communication over i2c with the module.
	Close() error

	// ManualCalibration(capacitance, location, disturber uint8) error

	// Reset the state of the module via PRESET_DEFAULT direct command register.
	InitializeDefaults() error

	// Enable disturber via MASK_DIST register.
	EnableDisturber() error

	// Disable disturber via MASK_DIST register.
	DisableDisturber() error

	// Set the source type of the IRQ pin interrupt via the DISP_LCO/DISP_SRCO/DISP_TRCO registers.
	SetIRQOutputSource(source IRQOutputSource) error

	// Set the internal capacitors capacitance in range from 0pF - 120pF via TUN_CAP register.
	SetTuningCapacitance(capacitance TuningCapacitance) error

	// Get the interrupt source type via the INT register.
	GetInterruptSource() (InterruptType, error)

	// Get estimated distance in KM of storm/latest lightning via the DISTANCE register. The value
	// "0" corresponds to "Storm ahead" and the "math.MaxInt" correspondes to "Out of range".
	GetLightningDistanceKm() (int, error)

	// Get the lightning strike energy via the S_LIG_MM/S_LIG_M/S_LIG_L registers.
	GetStrikeEnergy() (float64, error)

	// Set the environment tuning via the AFE_GB register.
	SetAnalogFrontEnd(model AnalogFrontEnd) error

	// Dump the value of registers from 0x00 to 0x08.
	DumpRegisters() ([9]uint8, error)

	// GetNoiseFloorLevel(level int) error

	// SetNoiseFloorLevel() (int, error)

	// GetWatchdogThreshold(threshold int) error

	// SetWatchdogThreshold() (int, error)

	// GetSpikeRejection() (int, error)

	// SetSpikeRejection(rejection int) error

	// Set the power up or down via the PWD register.
	PowerSwitch(power bool) error
}

// Create a instance of the AS3935 module from the provided device path and I2C address.
// All module functions are locking what allows to use the module in multiple goroutines.
func NewModule(device string, address int) (Module, error) {
	if len(device) == 0 {
		return nil, fmt.Errorf("as3935: the device file system name can not be empty")
	}

	return &module{
		DeviceFs:    device,
		Device:      nil,
		Address:     address,
		BufferRead:  make([]uint8, 1),
		BufferWrite: make([]uint8, 1),
		mu:          sync.Mutex{},
	}, nil
}

type module struct {
	DeviceFs    string
	Device      *i2c.Device
	Address     int
	BufferRead  []uint8
	BufferWrite []uint8
	mu          sync.Mutex
}

func (m *module) PowerSwitch(power bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !power {
		if err := m.RegWriteMasked(0x00, 0x01, 0x01); err != nil {
			return fmt.Errorf("as3935: failed to set the power down value to the register: %w", err)
		}

		return nil
	}

	if err := m.RegWriteMasked(0x00, 0x00, 0x01); err != nil {
		return fmt.Errorf("as3935: failed to set the power up value to the register: %w", err)
	}

	if err := m.RegWrite(0x3C, 0x96); err != nil {
		return fmt.Errorf("as3935: failed to set value to the calibration direct command register: %w", err)
	}

	if err := m.RegWriteMasked(0x08, uint8(SRCO), uint8(SRCO)); err != nil {
		return fmt.Errorf("as3935: failed to set the irq source up as powerup sequence to the register: %w", err)
	}

	time.Sleep(delayDuration)

	if err := m.RegWriteMasked(0x08, 0x00, uint8(SRCO)); err != nil {
		return fmt.Errorf("as3935: failed to set the irq source down as powerup sequence to the register: %w", err)
	}

	return nil
}

func (m *module) DumpRegisters() ([9]uint8, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		offset    uint8    = 0
		registers [9]uint8 = [9]uint8{}
		length    uint8    = uint8(len(registers))
		err       error    = nil
	)

	for offset < length {
		if registers[offset], err = m.RegRead(offset); err != nil {
			return [9]uint8{}, fmt.Errorf("as3935: failed to access one of the registers during the dump: %w", err)
		} else {
			offset += 1
		}
	}

	return registers, nil
}

func (m *module) DisableDisturber() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.RegWriteMasked(0x03, 0x00, 0x20); err != nil {
		return fmt.Errorf("as3935: failed to apply disable of disturber to register: %w", err)
	}

	return nil
}

func (m *module) EnableDisturber() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.RegWriteMasked(0x03, 0x20, 0x20); err != nil {
		return fmt.Errorf("as3935: failed to apply disable of disturber to register: %w", err)
	}

	return nil
}

func (m *module) GetInterruptSource() (InterruptType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	time.Sleep(delayDuration)

	register, err := m.RegRead(0x03)
	if err != nil {
		return NoResults, fmt.Errorf("as3935: failed to access the interrupt register: %w", err)
	}

	switch register & 0x0F {
	case uint8(NoResults):
		return NoResults, nil
	case uint8(NoiseLevelTooHigh):
		return NoiseLevelTooHigh, nil
	case uint8(DisturberDetected):
		return DisturberDetected, nil
	case uint8(LightningInterrupt):
		return LightningInterrupt, nil
	default:
		return NoResults, fmt.Errorf("as3935: invalid or corrupted interrupt data retrievef from register")
	}
}

func (m *module) GetLightningDistanceKm() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	register, err := m.RegRead(0x07)
	if err != nil {
		return 0, fmt.Errorf("as3935: failed to access the distance register: %w", err)
	}

	switch register & 0x3F {
	case 0x01:
		return 0, nil
	case 0x3F:
		return math.MaxInt, nil
	default:
		return int(register & 0x3F), nil
	}
}

func (m *module) GetStrikeEnergy() (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	registerL, err := m.RegRead(0x04)
	if err != nil {
		return 0, fmt.Errorf("as3935: failed to access l strike energy register: %w", err)
	}

	registerM, err := m.RegRead(0x05)
	if err != nil {
		return 0, fmt.Errorf("as3935: failed to access m strike energy register: %w", err)
	}

	registerMM, err := m.RegRead(0x06)
	if err != nil {
		return 0, fmt.Errorf("as3935: failed to access mm strike enregy register: %w", err)
	}

	// TODO: Verify if the formula is correct and is host endian agnostic
	var value uint32 = uint32(registerMM&0x1F) << 16
	value |= uint32(registerM) << 8
	value |= uint32(registerL)
	value /= 16777

	return float64(value) / 1000.0, nil
}

func (m *module) InitializeDefaults() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.RegWrite(0x3C, 0x96); err != nil {
		return fmt.Errorf("as3935: failed to apply initialize module defaults to reigster: %w", err)
	}

	return nil
}

func (m *module) SetAnalogFrontEnd(model AnalogFrontEnd) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch model {
	case Indoor, Outdoor:
	default:
		return fmt.Errorf("as3935: invalid analog frontend model specified")
	}

	if err := m.RegWriteMasked(0x00, uint8(model), 0x3E); err != nil {
		return fmt.Errorf("as3935: failed to apply the analog frontend to the register: %w", err)
	}

	return nil
}

func (m *module) SetIRQOutputSource(source IRQOutputSource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch source {
	case None, TRCO, SRCO, LCO:
	default:
		return fmt.Errorf("as3935: invalid IRQ output source specified")
	}

	if err := m.RegWriteMasked(0x08, uint8(source), 0xE0); err != nil {
		return fmt.Errorf("as3935: failed to apply irq output source to register: %w", err)
	}

	return nil
}

func (m *module) SetTuningCapacitance(capacitance TuningCapacitance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch capacitance {
	case TuningDiv16, TuningDiv32, TuningDiv64, TuningDiv128:
	default:
		return fmt.Errorf("as3935: invalid tuning capacitance value specified")
	}

	if err := m.RegWriteMasked(0x08, uint8(capacitance), 0x0F); err != nil {
		return fmt.Errorf("as3935: failed to apply the tuning capacitance to register: %w", err)
	}

	return nil
}

func (m *module) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Device == nil {
		return fmt.Errorf("as3935: the module is not connected")
	}

	defer func() {
		m.Device = nil
	}()

	if err := m.Device.Close(); err != nil {
		return fmt.Errorf("as3935: underlying i2c connection closing failure: %w", err)
	}

	return nil
}

func (m *module) Open() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Device != nil {
		return fmt.Errorf("as3935: the module is already connected")
	}

	devFs := &i2c.Devfs{
		Dev: m.DeviceFs,
	}

	dev, err := i2c.Open(devFs, m.Address)
	if err != nil {
		return fmt.Errorf("as3935: failed to open the connection to the module: %w", err)
	}

	m.Device = dev
	return nil
}

// Read a value from the register specified by the offset parameter.
func (m *module) RegRead(offset uint8) (uint8, error) {
	err := m.Device.ReadReg(offset, m.BufferRead)
	if err != nil {
		return 0x00, fmt.Errorf("as3935: failed to read the value at the given offset via i2c: %w", err)
	}

	return m.BufferRead[0], nil
}

// Write a value byte parameter to the register specified by the offset parameter.
func (m *module) RegWrite(offset, value uint8) error {
	m.BufferWrite[0] = value

	err := m.Device.WriteReg(offset, m.BufferWrite)
	if err != nil {
		return fmt.Errorf("as3935: failed to write the value at the given offset via i2c: %w", err)
	}

	return nil
}

// Replace bits from value parameter that are specified by "1" in the mask parameter to in register specified by the offset parameter.
func (m *module) RegWriteMasked(offset, value, mask uint8) error {
	register, err := m.RegRead(offset)
	if err != nil {
		return fmt.Errorf("as3935: failed to read the register for masked writing: %w", err)
	}

	register = (register & ^mask) | (value & mask)

	if err := m.RegWrite(offset, register); err != nil {
		return fmt.Errorf("as3935: failed to write the register for masked writing: %w", err)
	}

	return nil
}
