package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"
)

func AutoComplete(args []string, flags ...string) []string {
	if len(args) == 0 {
		return args
	}

	for _, it := range args {
		if len(it) > 0 && it[0] == '-' {
			for _, c := range it[1:] {
				if c < '0' || c > '9' {
					return args
				}
			}
		}
	}

	if len(args) > len(flags) {
		return args
	}

	ret := []string{}
	for i, arg := range args {
		flag := flags[i]
		ret = append(ret, "-" + flag + "=" + arg)
	}

	return ret
}

func NewFlagSet() *flag.FlagSet {
	return flag.NewFlagSet("", flag.ContinueOnError)
}

func ParseFlag(flag *flag.FlagSet, args []string, flags ...string) (help string, err error) {
	display := func() string {
		buf := bytes.NewBuffer(nil)

		if ArgsCount(flag) == 0 {
			fmt.Fprintln(buf, NoArgsNeed)
			return buf.String()
		}

		flag.SetOutput(buf)
		flag.PrintDefaults()
		flag.SetOutput(nil)

		fmt.Fprintf(buf, "\nshortcut:")
		for _, it := range flags {
			fmt.Fprint(buf, " <", it, ">")
		}
		fmt.Fprintln(buf)
		return buf.String()
	}
	help = display()

	if len(args) > 0 && (args[len(args) - 1] == "help" || args[len(args) - 1] == "?") {
		err = ErrWrongArgs
		return
	}

	args = AutoComplete(args, flags...)

	buf := bytes.NewBuffer(nil)
	flag.SetOutput(buf)
	err = flag.Parse(args)
	flag.SetOutput(nil)
	return
}

func ArgsCount(fs *flag.FlagSet) (count int) {
	fs.VisitAll(func(it *flag.Flag) {
		count += 1
	})
	return
}

func PopArg(name string, def string, args []string) (value string, repacked []string) {
	for i, arg := range args {
		if arg == "-"  + name && (i + 1 < len(args)) {
			value = args[i + 1]
			return value, append(args[:i], args[i + 2:]...)
		}
		if strings.HasPrefix(arg, "-" + name + "=") {
			value = args[i][len(name) + 2:]
			return value, append(args[:i], args[i + 1:]...)
		}
	}
	return def, args
}

func SplitArgs(args []string, target ...string) (result []string, repacked []string) {
	repacked = args
	for i, arg := range args {
		found := false
		for _, name := range target {
			if arg == "-"  + name && (i + 1 < len(args)) {
				value := args[i + 1]
				result = append(result, value)
				repacked = append(args[:i], args[i + 2:]...)
				found = true
				break
			}
			if strings.HasPrefix(arg, "-" + name + "=") {
				value := args[i][len(name) + 2:]
				result = append(result, value)
				repacked = append(args[:i], args[i + 1:]...)
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return
}

const NoArgsNeed = "no args need"

var ErrWrongArgs = errors.New("wrong args")
