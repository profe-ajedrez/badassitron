package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

type Options struct {
	Prec    int
	Process int
	NormUV  bool

	DetailTaxProcess DetailTaxProcessor
}

func (o *Options) DetailTaxProcessor() DetailTaxProcessor {
	return o.DetailTaxProcess
}

func (o *Options) Flow() int {
	return o.Process
}

func (o *Options) NormalizeUnitValue() bool {
	return o.NormUV
}

func (o *Options) UVNormalizer(n dec128.Dec128) dec128.Dec128 {
	return n
}

func (o *Options) ValueNormalizer(n dec128.Dec128) dec128.Dec128 {
	return n
}

func (o *Options) Scale() int {
	return o.Prec
}

func (o *Options) WithDetailTaxProcessor(dp DetailTaxProcessor) {
	o.DetailTaxProcess = dp
}

func (o *Options) WithNoUnitValueNormalization() {
	o.NormUV = false
}

func (o *Options) WithUnitValueNormalized() {
	o.NormUV = true
}

var _ CalculationConfiger = &Options{}
