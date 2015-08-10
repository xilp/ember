package main

import (
	"fmt"
	"os"
)

func (p *Log) SetLevel(level uint) (err error) {
	p.Level = level
	return
}

func (p *Log) GetLevel() (level uint, err error) {
	return p.Level, err
}

func (p *Log) SetPath(path string) (err error) {
	p.Path = path
	return
}

func (p *Log) GetPath() (path string, err error) {
	return p.Path, err
}

func (p *Log) Write(level uint,format string, args ...interface{}) (err error) {
	var x string
	if level < p.Level {
		return
	}
	if args == nil {
		x = fmt.Sprintf(format)
	} else {
		x = fmt.Sprintf(format, args)
	}
	p.cache = append(p.cache, x)
	p.cacheLine = p.cacheLine + 1
	if p.cacheLine > p.MaxCache {
		p.Flush()
	}
	return
}

func (p *Log) Flush() (err error) {
	for _, r := range p.cache {
		fmt.Printf("%s\n", r)
	}
	p.cacheLine = 0
	return
}

func NewLog() (log Log, err error) {
	log.Level = LOG_ERR
	log.Path = LOG_PATH
	log.AccessLog = log.Path + "/" + LOG_ACCESS
	log.ErrorLog = log.Path + "/" + LOG_ERROR
	log.MaxCache = LOG_MAX_CACHE 
	err = os.MkdirAll(log.Path, 0755)
	if err != nil {
		println(err.Error())
	}

	log.accessLog, err = os.OpenFile(log.AccessLog, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0640)
	if err != nil {
		return
	}
	log.errorLog, err = os.OpenFile(log.ErrorLog, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0640)
	if err != nil {
		return
	}

	return
}

func (p *Log) Init() (err error) {
	return
}

type Log struct {
	Level       uint   `json:"level" bson:"level"`
	Path        string `json:"path" bson:"path"`
	AccessLog	string `json:"accesslog" bson:"accesslog"`
	ErrorLog	string `json:"errorlog" bson:"errorlog"`
	MaxCache	int
	cache		[]string
	cacheLine int
	accessLog *os.File
	errorLog *os.File
}

const (
	LOG_STDERR    =  0
	LOG_EMERG     =  1
	LOG_ALERT     =  2
	LOG_CRIT      =  3
	LOG_ERR       =  4
	LOG_WARN      =  5
	LOG_NOTICE    =  6
	LOG_INFO      =  7
	LOG_DEBUG     =  8
	LOG_PATH      =  "log"
	LOG_ACCESS	  =  "access.log"
	LOG_ERROR	  =  "error.log"
	LOG_MAX_CACHE =  100
)
