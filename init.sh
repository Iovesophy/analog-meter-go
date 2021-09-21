#!/bin/sh -eux

go get -d -u gobot.io/x/gobot/...
go install gobot.io/x/gobot/platforms/firmata
go get github.com/gobuffalo/uuid
