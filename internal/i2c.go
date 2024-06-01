package internal

import (
	"fmt"
	"io"

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

const (
	ReadBufferSize  uint8 = 9
	WriteBufferSize uint8 = 1
)

// Create a new I2C device wrapper instance
func NewI2cDevice(device string, address int, debugOut io.Writer) (I2c, error) {
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
		BufferRead:  make([]uint8, ReadBufferSize),
		BufferWrite: make([]uint8, WriteBufferSize),
	}, nil
}

type i2cWrapper struct {
	DeviceFs    string
	Device      *i2c.Device
	Address     int
	BufferRead  []uint8
	BufferWrite []uint8
	DebugOut    io.Writer
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
	// TODO: The function is performing a workaround for the broken I2C reading in the AS3935 IC

	if offset >= ReadBufferSize {
		return 0x00, fmt.Errorf("as3935: the offset is out of the module register range")
	}

	if err := i.Device.ReadReg(0x00, i.BufferRead); err != nil {
		return 0x00, fmt.Errorf("as3935: failed to read the value at the given offset via i2c: %w", err)
	}

	// NOTE: Debug logging logic
	if i.DebugOut != nil {
		fmt.Fprintf(i.DebugOut, "[ Read ] Offset: 0x%02x:\n", offset)
		for regOffset, regValue := range i.BufferRead {
			if uint8(regOffset) == offset {
				fmt.Fprintf(i.DebugOut, "[%08b]", regValue)
			} else {
				fmt.Fprintf(i.DebugOut, " %08b ", regValue)
			}

			fmt.Fprintf(i.DebugOut, " ")
		}
		fmt.Fprintf(i.DebugOut, "\n")
	}

	return i.BufferRead[offset], nil
}

func (i *i2cWrapper) RegWrite(offset, value uint8) error {
	i.BufferWrite[0] = value

	// NOTE: Debug logging logic. Load registers into buffer to compare them
	if i.DebugOut != nil {
		if _, err := i.RegRead(offset); err != nil {
			return fmt.Errorf("as3935: failed to read the value at the given offset via i2c for logging purposes: %w", err)
		}
	}

	err := i.Device.WriteReg(offset, i.BufferWrite)
	if err != nil {
		return fmt.Errorf("as3935: failed to write the value at the given offset via i2c: %w", err)
	}

	if i.DebugOut != nil {
		fmt.Fprintf(i.DebugOut, "[ Write ] Value: 0x%02x Offset: 0x%02x:\n", value, offset)
		for regOffset, regValue := range i.BufferRead {
			if uint8(regOffset) == offset {
				fmt.Fprintf(i.DebugOut, "[%08b]", regValue)
			} else {
				fmt.Fprintf(i.DebugOut, " %08b ", regValue)
			}

			fmt.Fprintf(i.DebugOut, " ")
		}

		fmt.Fprintf(i.DebugOut, "\n")

		if _, err := i.RegRead(offset); err != nil {
			return fmt.Errorf("as3935: failed to read the value at the given offset via i2c for logging purposes: %w", err)
		}

		for regOffset, regValue := range i.BufferRead {
			if uint8(regOffset) == offset {
				fmt.Fprintf(i.DebugOut, "[%08b]", regValue)
			} else {
				fmt.Fprintf(i.DebugOut, " %08b ", regValue)
			}

			fmt.Fprintf(i.DebugOut, " ")
		}

		fmt.Fprintf(i.DebugOut, "\n")
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

	if i.DebugOut != nil {
		fmt.Fprintf(i.DebugOut, "[ Write Masked ] Value: 0x%02x Mask: 0x%02x Offset: 0x%02x:\n", value, mask, offset)
		for regOffset, regValue := range i.BufferRead {
			if uint8(regOffset) == offset {
				fmt.Fprintf(i.DebugOut, "[%08b]", regValue)
			} else {
				fmt.Fprintf(i.DebugOut, " %08b ", regValue)
			}

			fmt.Fprintf(i.DebugOut, " ")
		}

		fmt.Fprintf(i.DebugOut, "\n")

		if _, err := i.RegRead(offset); err != nil {
			return fmt.Errorf("as3935: failed to read the value at the given offset via i2c for logging purposes: %w", err)
		}

		for regOffset, regValue := range i.BufferRead {
			if uint8(regOffset) == offset {
				fmt.Fprintf(i.DebugOut, "[%08b]", regValue)
			} else {
				fmt.Fprintf(i.DebugOut, " %08b ", regValue)
			}

			fmt.Fprintf(i.DebugOut, " ")
		}

		fmt.Fprintf(i.DebugOut, "\n")
	}

	return nil
}
