package nlogn

type LogLevel uint8

const (
	NONE LogLevel = iota
	CRITICAL
	ERROR
	WARNING
	INFO
	DEBUG
	TRACE
)
