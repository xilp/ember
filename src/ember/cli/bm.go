package cli

import (
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"
)

type Invoke func() error

var ErrBmExit = errors.New("bm exit")

type BenchMark struct {
	stopping bool
}

func NewBenchMark() *BenchMark {
	return &BenchMark{}
}

func (p *BenchMark) Run(invoke Invoke, times int, conc int, ignore bool) (ns int64, err error) {
	if times <= 0 || conc <= 0 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(times)

	locker := sync.Mutex{}
	errLocker := sync.Mutex{}
	counter := times
	now := time.Now().UnixNano()

	for i := 0; i < conc; i++ {
		go func() {
			for {
				locker.Lock()
				counter -= 1
				t := counter
				locker.Unlock()
				if t < 0 {
					break
				}

				if !p.stopping {
					err1 := invoke()
					if err1 == ErrBmExit {
						p.stopping = true
					} else if err1 != nil && !ignore {
						errLocker.Lock()
						if err == nil {
							err = err1
						}
						errLocker.Unlock()
						p.stopping = true
					}
				}
				wg.Done()
			}
		}()
	}

	wg.Wait()
	ns = time.Now().UnixNano() - now
	return
}

func (p *BenchMark) Dbm(args []string, tester func(s int, mute bool)error) {
	flag := flag.NewFlagSet("bench mark", flag.ContinueOnError)
	times := flag.Int("t", 1, "repeat times")
	min := flag.Int("min", 0, "min data size")
	max := flag.Int("max", 1024 * 1024 * 4, "max data size")
	ignore := flag.Bool("f", false, "ignore errors")
	conc := flag.Int("c", 1, "concurrency")
	mute := flag.Bool("m", false, "do not show size when use random size")

	ParseFlag(flag, args, "t", "min", "max", "f", "c", "m")

	io := int64(0)
	invoke := func() (error) {
		size := Rand(*min, *max)
		err := tester(size, *mute)
		if err == nil {
			io += int64(size)
		}
		return err
	}
	ns, err := p.Run(invoke, *times, *conc, *ignore)

	if err != nil {
		fmt.Print("error:     ", err.Error(), "\n")
		return
	}

	if !*mute {
		fmt.Println()
	}

	fmt.Print("size:      ", *min, " ~ ", *max, "\n")
	fmt.Print("times:     ", *times, "\n")
	fmt.Print("elapsed:   ", Nms(ns, 1), "\n")
	if ns != 0 {
		fmt.Print("iotp:      ", Bkmg(Tps(io, ns), 1), "\n")
		fmt.Print("iops:      ", Tps(int64(*times), ns), "\n")
	}
	fmt.Println()
}
