package measure

import (
	"testing"
)

func TestSpecData(t *testing.T) {
	d := NewSpecData()
	d.Record(1)
	d.Record(9)
	d.Record(5)
	if d.Max != 9 || d.Min != 1 || d.Sum != 15 || d.Count != 3 {
		t.Fatal("wrong", d)
	}
}

func TestSpanData(t *testing.T) {
	d := NewSpanData()
	d.Record("a", 1)
	d.Record("a", 9)
	d.Record("a", 5)
	d.Record("b", 15)

	c := d.Copy()
	c.Record("b", 7)

	a := d.Data["a"]
	if a == nil || a.Max != 9 || a.Min != 1 || a.Sum != 15 || a.Count != 3 {
		t.Fatal("wrong", d)
	}

	d.Clear()
	if len(d.Data) != 0 || d.Time != 0 {
		t.Fatal("wrong", d)
	}

	b := c.Data["b"]
	if b == nil || b.Max != 15 || b.Min != 7 || b.Sum != 22 || b.Count != 2 {
		t.Fatal("wrong", c, b)
	}
}

func TestMeasureData(t *testing.T) {
	d := NewMeasureData(2)
	d.Record(100, "a", 1)
	d.Record(100, "a", 9)
	d.Record(100, "a", 5)
	d.Record(200, "a", 5)

	if d[len(d) - 1].Time != d.LastTime() {
		t.Fatal("wrong", d)
	}

	last := d[len(d) - 1].Data["a"]
	if last.Count != 1 {
		t.Fatal("wrong", last)
	}

	c := d.Copy()
	c.Record(200, "a", 0)

	last = d[len(d) - 1].Data["a"]
	if last.Count != 1 {
		t.Fatal("wrong", last)
	}

	last = c[len(c) - 1].Data["a"]
	if last.Count != 2 {
		t.Fatal("wrong", c)
	}

	//func (p *MeasureData) Padding(count int) {

	//func (p *MeasureData) Merge(x *MeasureData) MeasureData {
	//func (p *MeasureData) After(time int64) MeasureData {
}
