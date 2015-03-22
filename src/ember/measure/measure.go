package spy

import (
	"time"
	"sync"
)

// TODO: sync
func (p *Measure) Dump(name string, val int64) []*SpanData {
	p.locker.Lock()
	defer p.locker.Unlock()
	result := make([]*SpanData, len(p.data))
	for i, it := range p.data {
		result[i] = it.Copy()
	}
	return result
}

func (p *Measure) Record(name string, value int64) {
	p.locker.Lock()
	defer p.locker.Unlock()

	now := time.Now().UnixNano()
	now = now / p.interval * p.interval

	last := p.data[len(p.data) - 1]
	if last.Time > now {
		padding := int((now - last.Time) / p.interval)
		for i := 0; i < padding; i++ {
			p.data = append(p.data, NewSpanData())
		}
		p.data = p.data[padding:]
	}
	if last.Time == 0 {
		last.Time = now
	}

	if last.Time == now {
		last.Merge(name, value)
	} else if last.Time > now {
		panic("Measure.Record: unexpect")
	} else {
		for i := len(p.data) - 1; i >= 0; i-- {
			if p.data[i].Time != now {
				continue
			}
			p.data[i].Merge(name, value)
		}
	}
}

func NewMeasure(interval time.Duration, keep time.Duration) *Measure {
	count := keep / interval
	p := &Measure {
		interval: int64(interval),
		data: make([]*SpanData, count),
	}
	for i, _ := range p.data {
		p.data[i] = NewSpanData()
	}
	return p
}

type Measure struct {
	interval int64
	data []*SpanData
	locker sync.Mutex
}
