package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/firmata"
	"gobot.io/x/gobot/platforms/keyboard"
)

const (
	MeterMax  uint8         = 165
	MeterMin                = 15
	MeterWait time.Duration = 200
)

type device struct {
	keys       *keyboard.Driver
	keyflag    string
	ledGreen   *gpio.LedDriver
	ledBlue    *gpio.LedDriver
	ledYellow  *gpio.LedDriver
	servoMotor *gpio.ServoDriver
	angleBuf   uint8
}

func (c *device) deviceSettings(firmataAdaptor *firmata.Adaptor) {
	c.servoMotor = gpio.NewServoDriver(firmataAdaptor, "5")
	c.ledYellow = gpio.NewLedDriver(firmataAdaptor, "3")
	c.ledGreen = gpio.NewLedDriver(firmataAdaptor, "11")
	c.ledBlue = gpio.NewLedDriver(firmataAdaptor, "13")
}

func (c *device) initMotion() {
	c.angleBuf = 165
	c.servoMotor.Move(c.angleBuf)
	for i := 0; i < 5; i++ {
		c.ledBlue.Toggle()
		c.ledYellow.Toggle()
		c.ledGreen.Toggle()
		time.Sleep(time.Millisecond * 100)
	}
}

func (c *device) subLoop() {
	for {
		if c.keyflag == "cpu" {
			p, err := cpu.Percent(0, false)
			if err != nil {
				log.Fatal(err)
			}
			if p[0] != 0 {
				angleRaw := uint8(p[0] / 100 * float64(MeterMax-MeterMin))
				angle := MeterMax - angleRaw
				if c.angleBuf <= angle {
					for i := c.angleBuf; i < angle; i++ {
						c.servoMotor.Move(i)
						time.Sleep(time.Millisecond * 50)
					}
				} else if c.angleBuf >= angle {
					for i := c.angleBuf; i > angle; i-- {
						c.servoMotor.Move(i)
						time.Sleep(time.Millisecond * 50)
					}
				}
				c.angleBuf = angle
			}
			time.Sleep(time.Millisecond * MeterWait)
		} else if c.keyflag == "mem" {
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Fatal(err)
			}
			if m.UsedPercent != 0 {
				angleRaw := uint8(m.UsedPercent / 100 * float64(MeterMax-MeterMin))
				c.servoMotor.Move(MeterMax - angleRaw)
			}
			time.Sleep(time.Millisecond * MeterWait)
		} else if c.keyflag == "disk" {
			d, err := disk.Usage("/Volumes")
			if err != nil {
				log.Fatal(err)
			}
			if d.UsedPercent != 0 {
				angleRaw := uint8(d.UsedPercent / 100 * float64(MeterMax-MeterMin))
				c.servoMotor.Move(MeterMax - angleRaw)
			}
			time.Sleep(time.Millisecond * MeterWait)
		}
	}
}

func (c *device) mainLoop() {
	firmataAdaptor := firmata.NewAdaptor("/dev/tty.usbmodem142101")
	c.deviceSettings(firmataAdaptor)
	c.keys = keyboard.NewDriver()
	c.keys.On(keyboard.Key, func(keydata interface{}) {
		key := keydata.(keyboard.KeyEvent)
		c.initMotion()
		if key.Key == keyboard.P {
			c.keyflag = "cpu"
			c.ledBlue.Off()
			c.ledYellow.Off()
			c.ledGreen.On()
		} else if key.Key == keyboard.M {
			c.keyflag = "mem"
			c.ledBlue.Off()
			c.ledYellow.On()
			c.ledGreen.Off()
		} else if key.Key == keyboard.D {
			c.keyflag = "disk"
			c.ledBlue.On()
			c.ledYellow.Off()
			c.ledGreen.Off()
		} else if key.Key >= 97 && key.Key <= 122 {
			c.keyflag = "unknownkey"
			c.ledBlue.Off()
			c.ledYellow.Off()
			c.ledGreen.Off()
			c.servoMotor.Move(165)
		}
		fmt.Println(c.keyflag)
	})
	robot := gobot.NewRobot("bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{
			c.keys,
			c.servoMotor,
			c.ledBlue,
			c.ledYellow,
			c.ledGreen,
		},
		c.keys, //set workspace
	)
	robot.Start()
}

func main() {
	c := device{}
	c.keyflag = "init"
	go c.subLoop()
	c.mainLoop()
}
