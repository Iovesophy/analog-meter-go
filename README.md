# analog-meter-go

Only support on Darwin , if you want to use on any OS , please setup yourself .

## usage

- setup Firmata in arduino (StandardFirmata)

>  Firmata is a generic protocol for communicating with microcontrollers
> from software on a host computer. It is intended to work with
> any host computer software package.
> To download a host software package, please click on the following link
> to open the list of Firmata client libraries in your default browser.

https://github.com/firmata/arduino#firmata-client-libraries

- create hardware

### Parts list and PIN

```Go
func (c *device) deviceSettings(firmataAdaptor *firmata.Adaptor) {
	c.servoMotor = gpio.NewServoDriver(firmataAdaptor, "5")
	c.ledYellow = gpio.NewLedDriver(firmataAdaptor, "3")
	c.ledGreen = gpio.NewLedDriver(firmataAdaptor, "11")
	c.ledBlue = gpio.NewLedDriver(firmataAdaptor, "13")
}
```

- connect arduino to your PC

- check tty

```bash
ls -l /dev/tty.*
```
and setting arduino tty path .


- build and run

```bash
$ go build main.go
$ ./main
```

Authors
kazuya yuda.

Copyright (c) 2021 Kazuya yuda.
