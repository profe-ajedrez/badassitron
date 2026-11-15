package internal

import "testing"

func TestImporte(t *testing.T) {
	for _, tc := range []struct {
		name  string
		base  float64
		scale int
		want  float64
	}{
		{name: "zero scale", base: 12, scale: 0, want: 11.5},
		{name: "positive scale", base: 100, scale: 1, want: 50},
		{name: "negative scale", base: 1, scale: -1, want: 0.995},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Importe(tc.base, tc.scale)
			if got != tc.want {
				t.Fatalf("Importe(%v, %d) = %v, want %v", tc.base, tc.scale, got, tc.want)
			}
		})
	}
}

func TestNumDecPlaces(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  int
	}{
		{0, 0},
		{123, 0},
		{-123.25, 2},
		{0.001, 3},
		{1.000001, 6},
	} {
		if got := NumDecPlaces(tc.value); got != tc.want {
			t.Errorf("NumDecPlaces(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
