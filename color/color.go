package color

import (
	colorful "github.com/lucasb-eyer/go-colorful"
)

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

// This is the meat of the gradient computation. It returns a HCL-blend between
// the two colors around `t`.
// Note: It relies heavily on the fact that the gradient keypoints are sorted.
func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self)-1; i++ {
		c1 := self[i]
		c2 := self[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self[len(self)-1].Col
}

// This is a very nice thing Golang forces you to do!
// It is necessary so that we can write out the literal of the colortable below.
func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}

func Calc(val float64) string {
	keypoints := GradientTable{
		{MustParseHex("#0000FF"), -50},
		{MustParseHex("#0000FF"), -20},
		{MustParseHex("#00FF00"), 0},
		{MustParseHex("#FF0000"), 20},
		{MustParseHex("#FF0000"), 50},
	}

	c := keypoints.GetInterpolatedColorFor(val)
	return c.Hex()
}
