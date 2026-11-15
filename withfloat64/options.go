package withfloat64

// Options is the concrete configuration type for the staged calculation.
type Options struct {
	Prec    int
	Process int
	NormUV  bool

	DetailTaxProcess DetailTaxProcessor
}

// DetailTaxProcessor returns the configured tax detail processor.
func (o *Options) DetailTaxProcessor() DetailTaxProcessor {
	return o.DetailTaxProcess
}

// Flow returns the configured calculation flow.
func (o *Options) Flow() int {
	return o.Process
}

// NormalizeUnitValue reports whether unit value normalization is enabled.
func (o *Options) NormalizeUnitValue() bool {
	return o.NormUV
}

// UVNormalizer returns the unit value using the configured normalizer.
func (o *Options) UVNormalizer(n float64) float64 {
	return n
}

// ValueNormalizer returns the value using the configured normalizer.
func (o *Options) ValueNormalizer(n float64) float64 {
	return n
}

// Scale returns the configured decimal scale.
func (o *Options) Scale() int {
	return o.Prec
}

// WithDetailTaxProcessor sets the tax detail processor used by calculation handlers.
func (o *Options) WithDetailTaxProcessor(dp DetailTaxProcessor) {
	o.DetailTaxProcess = dp
}

// WithNoUnitValueNormalization disables unit value normalization.
func (o *Options) WithNoUnitValueNormalization() {
	o.NormUV = false
}

// WithUnitValueNormalized enables unit value normalization.
func (o *Options) WithUnitValueNormalized() {
	o.NormUV = true
}

var _ CalculationConfiger = &Options{}
