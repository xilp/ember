package spy

import (
	"time"
)

func (p *SpanData) Merge(name string, value int64) {
	if _, ok := p.Data[name]; !ok {
		p.Data[name] = &SpecData{}
	}
	p.Data[name].Merge(value)
}

func (p *SpanData) Clear() {
	p.Time = 0
	p.Data = make(map[string]*SpecData)
}

func (p *SpanData) Copy() *SpanData {
	result := &SpanData {
		p.Time,
		make(map[string]*SpecData),
	}
	for k, v := range p.Data {
		var cp = *v
		result.Data[k] = &cp
	}
	return result
}

func NewSpanData() *SpanData {
	return &SpanData {
		time.Now().UnixNano(),
		make(map[string]*SpecData),
	}
}

type SpanData struct {
	Time int64
	Data map[string]*SpecData
}

func (p *SpecData) Merge(value int64) {
	p.Max = Max(p.Max, value)
	p.Min = Min(p.Min, value)
	p.Sum = Sum(p.Sum, value)
	p.Count = Count(p.Count, value)
}

type SpecData struct {
	Max, Min, Sum, Count int64
}
