module doba.com/goproxy_demo

go 1.17

require (
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
)

require golang.org/x/text v0.3.3 // indirect

replace golang.org/x/net => ../../golang.org/x/net
