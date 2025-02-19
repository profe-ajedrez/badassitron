package badassitron

// AppliesTo represents when to apply the value of something, when is by each unit or for the entirety of the line
type AppliesTo int8

// Origin tell from which value we should start making the calculations to obtain other values

const (
	Unit = AppliesTo(0)
	Line = AppliesTo(1)
)
