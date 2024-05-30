# AS3935 (Go)
[![Go Reference](https://pkg.go.dev/badge/github.com/Krzysztofz01/as3935-go.svg)](https://pkg.go.dev/github.com/Krzysztofz01/as3935-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Krzysztofz01/as3935-go)](https://goreportcard.com/report/github.com/Krzysztofz01/as3935-go)
![GitHub](https://img.shields.io/github/license/Krzysztofz01/as3935-go)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/Krzysztofz01/as3935-go?include_prereleases)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Krzysztofz01/as3935-go)

The purpose of the library is to provide a simple interface for interacting with the AS3935 module, which is capable of detecting lightning strikes, in the Go language. The library uses a module that allows to handle communication via i2c. The library allows to make changes and read values from to the module's registers, which are described in the documentation.

## Sources
The library is an alternative to implementations such as (which were very helpful to udnerstand the communication and logic of the module): 
- [github.com/DFRobot/DFRobot_AS3935 (Python implementation)](https://github.com/DFRobot/DFRobot_AS3935/blob/master/python/raspberrypi/DFRobot_AS3935_Lib.py)
- [github.com/DFRobot/DFRobot_AS3935 (C++ implementation)](https://github.com/DFRobot/DFRobot_AS3935/blob/master/DFRobot_AS3935_I2C.cpp)

Documentation:
- [AS3935](https://raw.githubusercontent.com/DFRobot/Wiki/SEN0290/DFRobot_SEN0290/res/AS3935_Franklin%20Lightning%20Sensor%20IC.pdf)
- [MA5532-AE](https://raw.githubusercontent.com/DFRobot/Wiki/SEN0290/DFRobot_SEN0290/res/Coilcraft%20MA5532-AE.pdf)
- [Module manufacturer page](https://www.dfrobot.com/product-1828.html)

## Example
```go
package main

func main() {
    fmt.Printf("TODO\n")
}
```