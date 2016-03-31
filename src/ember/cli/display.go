package cli

import (
	"strconv"
	"strings"
)

func Istr(i int, width int) string {
	return Lstr(int64(i), width)
}

func Ustr(u uint32, width int) string {
	return Lstr(int64(u), width)
}

func Lstr(l int64, width int) string {
	s := strconv.FormatInt(l, 10)
	if len(s) < width {
		s = strings.Repeat(" ", width-len(s)) + s
	}
	return Lpad(s, width)
}

func Lustr(l uint64, width int) string {
	s := strconv.FormatUint(l, 10)
	if len(s) < width {
		s = strings.Repeat(" ", width-len(s)) + s
	}
	return Lpad(s, width)
}

func Lpad(s string, width int) string {
	if len(s) < width {
		s = strings.Repeat(" ", width-len(s)) + s
	}
	return s
}

func Rpad(s string, width int) string {
	if len(s) < width {
		s += strings.Repeat(" ", width-len(s))
	}
	return s
}

func shrink(number int64, depth int, max int64, step int) (n int64, d int) {
	for number > max {
		number = number / int64(step)
		depth += 1
	}
	return number, depth
}

func Bkmg(number int64, width int) string {
	n, d := shrink(number, 0, 10000, 1024)
	units := []string{"B", "K", "M", "G", "T", "P"}
	return Istr(int(n), width) + units[d]
}

func Kmg(number int64, width int) string {
	n, d := shrink(number, 0, 10000, 1024)
	units := []string{" ", "K", "M", "G", "T", "P"}
	return Istr(int(n), width) + units[d]
}

func Ms(nano int64, width int) string {
	return Istr(int(nano / 1000 / 1000), width) + "MS"
}

func Nms(nano int64, width int) string {
	ms := int(nano / 1000 / 1000)
	if ms < 10000 {
		return Istr(ms, width) + "MS"
	}
	return Istr(ms / 1000, width) + "S"
}

func Tps(val int64, ns int64) int64 {
	if ns == 0 || val == 0 {
		return 0
	}
	return val * 10000 / (ns / 100000)
}
