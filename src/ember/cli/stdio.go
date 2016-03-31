package cli

import (
	"bufio"
	"errors"
	"io"
	"fmt"
	"os"
)

func InLn(bufSize int, fun func([]byte)) error {
	r := bufio.NewReaderSize(os.Stdin, bufSize)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if err != io.EOF {
				Errln(err.Error())
				return err
			} else {
				return nil
			}
		}
		fun(line)
	}
	return nil
}

func In(bufSize int, fun func([]byte)) error {
	r := bufio.NewReaderSize(os.Stdin, bufSize)
	for {
		line, prefix, err := r.ReadLine()
		if err != nil {
			if err != io.EOF {
				Errln(err.Error())
				return err
			} else {
				return nil
			}
		}
		if prefix {
			return errors.New("line too long")
		}
		fun(line)
	}
	return nil
}

func Err(msg ...interface{}) {
	os.Stderr.WriteString(fmt.Sprint(msg...))
}

func Errln(msg ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintln(msg...))
}

func Check(err error) {
	if err == nil {
		return
	}
	Errln(err.Error())
	os.Exit(1)
}
