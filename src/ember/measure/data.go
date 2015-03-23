package spy

import (
	"time"
)

func (p *MeasureData) After(time int64) MeasureData {
	data := *p
	i := len(data) - 1
	if time > data[i].Time {
		return NewMeasureData(0)
	}

	for ; i >= 0; i-- {
		if data[i].Time == time {
			break
		}
	}

	after := MeasureData(data[i:])
	return after.Copy()
}

func (p *MeasureData) Copy() MeasureData {
	result := NewMeasureData(len(*p))
	for i, it := range (*p) {
		result[i] = it.Copy()
	}
	return result
}

func (p *MeasureData) LastTime() int64 {
	return (*p)[len(*p) - 1].Time
}

func (p *MeasureData) Padding(count int) {
	data := *p
	if count >= len(data) {
		*p = NewMeasureData(len(data))
	} else {
		for i := 0; i < count; i++ {
			*p = append(*p, NewSpanData())
		}
		*p= (*p)[count:]
	}
}

func (p *MeasureData) Merge(time int64, name string, value int64) {
	data := *p
	last := data[len(data) - 1]

	if last.Time == 0 {
		last.Time = time
	}

	for i := len(data) - 1; i >= 0; i-- {
		if data[i].Time != time {
			continue
		}
		data[i].Merge(name, value)
		return
	}

	panic("Measure.Record: unexpect")
}

func NewMeasureData(count int) MeasureData {
	data := make([]*SpanData, count)
	for i, _ := range data {
		data[i] = NewSpanData()
	}
	return data
}

type MeasureData []*SpanData

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
