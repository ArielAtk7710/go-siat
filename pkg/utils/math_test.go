package utils

import (
	"math"
	"testing"
)

func TestRound(t *testing.T) {
	cases := []struct {
		name     string
		v        float64
		decimals int
		want     float64
	}{
		{"dos decimales trunca", 123.456, 2, 123.46},
		{"cinco decimales", 1.123456789, 5, 1.12346},
		{"cero decimales", 7.5, 0, 8},
		{"negativo", -123.456, 2, -123.46},
		{"ya exacto", 10.5, 2, 10.5},
		{"cero", 0, 2, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Round(c.v, c.decimals)
			if got != c.want {
				t.Errorf("Round(%v, %d) = %v, quería %v", c.v, c.decimals, got, c.want)
			}
		})
	}
}

func TestRound_PanicConNaN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba panic con NaN, no ocurrió")
		}
	}()
	Round(math.NaN(), 2)
}

func TestRound_PanicConInfinito(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba panic con +Inf, no ocurrió")
		}
	}()
	Round(math.Inf(1), 2)
}

func TestRound_PanicConInfinitoNegativo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("se esperaba panic con -Inf, no ocurrió")
		}
	}()
	Round(math.Inf(-1), 2)
}
