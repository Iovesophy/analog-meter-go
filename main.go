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
	MeterMax uint8 = 165
	MeterMin       = 15
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
	c.ledBlue = gpio.NewLedDriver(firmataAdaptor, "11")
	c.ledGreen = gpio.NewLedDriver(firmataAdaptor, "13")
}

func (c *device) initMotion() {
	c.angleBuf = MeterMax
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
		time.Sleep(time.Second)
		if c.keyflag == "cpu" {
			p, err := cpu.Percent(0, false)
			if err != nil {
				log.Fatal(err)
			}
			angleRaw := calcAngleRaw(p[0])
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
		} else if c.keyflag == "mem" {
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Fatal(err)
			}
			angleRaw := calcAngleRaw(m.UsedPercent)
			c.servoMotor.Move(MeterMax - angleRaw)
		} else if c.keyflag == "disk" {
			d, err := disk.Usage("/Volumes")
			if err != nil {
				log.Fatal(err)
			}
			angleRaw := calcAngleRaw(d.UsedPercent)
			c.servoMotor.Move(MeterMax - angleRaw)
		}
	}
}

func calcAngleRaw(d float64) uint8 {
	return uint8(d / 100 * float64(MeterMax-MeterMin))
}

func (c *device) mainLoop() {
	firmataAdaptor := firmata.NewAdaptor("/dev/tty.usbmodem141101")
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
			c.servoMotor.Move(MeterMax)
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
	fmt.Println("start")
	c := device{}
	c.keyflag = "init"
	go c.subLoop()
	c.mainLoop()
}
