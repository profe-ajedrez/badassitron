package withfloat64

import (
	"encoding/json"
)

type CalculationConfig struct {
	scale                      int  // Scale
	flow                       int  // Flow
	scaleToNormalizedUnitValue int  // Scale to normalized unit value
	normalizeUnitValue         bool // Normalize unit value
}

func (cc *CalculationConfig) Scale() int {
	return cc.scale
}

func (cc *CalculationConfig) Flow() int {
	return cc.flow
}

func (cc *CalculationConfig) ScaleToNormalizedUnitValue() int {
	return cc.scaleToNormalizedUnitValue
}

func (cc *CalculationConfig) NormalizeUnitValue() bool {
	return cc.normalizeUnitValue
}

func (cc *CalculationConfig) WithScaleToNormalizedUnitValue(scale int) {
	cc.scaleToNormalizedUnitValue = scale
}

func (cc *CalculationConfig) WithUnitValueNormalized() {
	cc.normalizeUnitValue = true
}

func (cc *CalculationConfig) WithNoUnitValueNormalization() {
	cc.normalizeUnitValue = false
}

type Input struct {
	UV      float64     // Unit Value
	GT      float64     // Gross Total
	QTY     float64     // Quantity
	Disc    float64     // Discount
	TaxList []*InputTax // Taxes
}

func (i *Input) UnitValue() float64 {
	return i.UV
}

func (i *Input) GrossTotal() float64 {
	return i.GT
}

func (i *Input) Qty() float64 {
	return i.QTY
}

func (i *Input) Discount() float64 {
	return i.Disc
}

func (i *Input) Taxes() []TaxInformer {
	if i.TaxList == nil {
		return nil
	}
	taxes := make([]TaxInformer, len(i.TaxList))
	for idx, tax := range i.TaxList {
		taxes[idx] = tax
	}
	return taxes
}

func (i *Input) WithUnitValue(uv float64) {
	i.UV = uv
}

func (i *Input) WithGrossTotal(gt float64) {
	i.GT = gt
}

func (i *Input) SetDiscToZero() {
	i.Disc = 0
}

func (i *Input) SetDiscToHundred() {
	i.Disc = 100
}

type InputTax struct {
	CodeValue string  // Tax code
	NameValue string  // Tax name
	V         float64 // Tax value
	Id        int     // Tax ID
	Typee     Type    // Tax type
	Stagee    Stage   // Tax stage
}

// Stage implements TaxInformer.
func (it *InputTax) Stage() Stage {
	return it.Stagee
}

// Value implements TaxInformer.
func (it *InputTax) Value() float64 {
	return it.V
}

func (it *InputTax) ID() int {
	return it.Id
}

func (it *InputTax) Name() string {
	return it.NameValue
}

func (it *InputTax) Code() string {
	return it.CodeValue
}

func (it *InputTax) Type() Type {
	return it.Typee
}

func (it *InputTax) String() string {
	js, _ := json.Marshal(it)
	return string(js)
}

var _ Enterable = (*Input)(nil)
var _ TaxInformer = (*InputTax)(nil)
