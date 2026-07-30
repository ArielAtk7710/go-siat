package datatype

import (
	"math"
	"testing"
)

func TestFloat64Round(t *testing.T) {
	cases := []struct {
		name     string
		v        float64
		decimals int
		want     float64
	}{
		{"dos decimales trunca", 123.456, 2, 123.46},
		{"cinco decimales", 1.123456789, 5, 1.12346},
		{"diez decimales", 0.12345678901234, 10, 0.1234567890},
		{"negativo", -123.456, 2, -123.46},
		{"cero", 0, 2, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Float64Round(c.v, c.decimals)
			if got != c.want {
				t.Errorf("Float64Round(%v, %d) = %v, quería %v", c.v, c.decimals, got, c.want)
			}
		})
	}
}

func TestFloat64Round_PanicConNaN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba panic con NaN, no ocurrió")
		}
	}()
	Float64Round(math.NaN(), 2)
}

func TestFloat64Round_PanicConInfinito(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba panic con +Inf, no ocurrió")
		}
	}()
	Float64Round(math.Inf(1), 2)
}
