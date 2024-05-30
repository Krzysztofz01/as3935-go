package internal

import (
	"fmt"

	"golang.org/x/exp/io/i2c"
)

type I2c interface {
	// Open the connection to the i2c device
	Open() error

	// Close the underlying i2c device connection
	Close() error

	// Read a value from the register specified by the offset parameter.
	RegRead(offset uint8) (uint8, error)

	// Write a value byte parameter to the register specified by the offset parameter.
	RegWrite(offset, value uint8) error

	// Replace bits from value parameter that are specified by "1" in the mask parameter to in register specified by the offset parameter.
	RegWriteMasked(offset, value, mask uint8) error
}

// Create a new I2C device wrapper instance
func NewI2cDevice(device string, address int) (I2c, error) {
	if len(device) == 0 {
		return nil, fmt.Errorf("as3935: invalid i2c device specified")
	}

	if address < 0 {
		return nil, fmt.Errorf("as3935: invalid i2c address specified")
	}

	return &i2cWrapper{
		DeviceFs:    device,
		Device:      nil,
		Address:     address,
		BufferRead:  make([]uint8, 1),
		BufferWrite: make([]uint8, 1),
	}, nil
}

type i2cWrapper struct {
	DeviceFs    string
	Device      *i2c.Device
	Address     int
	BufferRead  []uint8
	BufferWrite []uint8
}

func (i *i2cWrapper) Close() error {
	if i.Device == nil {
		return fmt.Errorf("as3935: the module is not connected")
	}

	defer func() {
		i.Device = nil
	}()

	if err := i.Device.Close(); err != nil {
		return fmt.Errorf("as3935: underlying i2c connection closing failure: %w", err)
	}

	return nil
}

func (i *i2cWrapper) Open() error {
	if i.Device != nil {
		return fmt.Errorf("as3935: the module is already connected")
	}

	devFs := &i2c.Devfs{
		Dev: i.DeviceFs,
	}

	dev, err := i2c.Open(devFs, i.Address)
	if err != nil {
		return fmt.Errorf("as3935: failed to open the connection to the module: %w", err)
	}

	i.Device = dev
	return nil
}

func (i *i2cWrapper) RegRead(offset uint8) (uint8, error) {
	err := i.Device.ReadReg(offset, i.BufferRead)
	if err != nil {
		return 0x00, fmt.Errorf("as3935: failed to read the value at the given offset via i2c: %w", err)
	}

	return i.BufferRead[0], nil
}

func (i *i2cWrapper) RegWrite(offset, value uint8) error {
	i.BufferWrite[0] = value

	err := i.Device.WriteReg(offset, i.BufferWrite)
	if err != nil {
		return fmt.Errorf("as3935: failed to write the value at the given offset via i2c: %w", err)
	}

	return nil
}

func (i *i2cWrapper) RegWriteMasked(offset, value, mask uint8) error {
	register, err := i.RegRead(offset)
	if err != nil {
		return fmt.Errorf("as3935: failed to read the register for masked writing: %w", err)
	}

	register = (register & ^mask) | (value & mask)

	if err := i.RegWrite(offset, register); err != nil {
		return fmt.Errorf("as3935: failed to write the register for masked writing: %w", err)
	}

	return nil
}
