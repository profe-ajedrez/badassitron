package withdec128

import (
	"encoding/json"

	"github.com/profe-ajedrez/badassitron/dec128"
)

// CalculationConfig stores the calculation settings used by older workflows.
type CalculationConfig struct {
	scale                      int  // Scale
	flow                       int  // Flow
	scaleToNormalizedUnitValue int  // Scale to normalized unit value
	normalizeUnitValue         bool // Normalize unit value
}

// Scale returns the configured decimal scale.
func (cc *CalculationConfig) Scale() int {
	return cc.scale
}

// Flow returns the configured calculation flow.
func (cc *CalculationConfig) Flow() int {
	return cc.flow
}

// ScaleToNormalizedUnitValue returns the target scale for unit value normalization.
func (cc *CalculationConfig) ScaleToNormalizedUnitValue() int {
	return cc.scaleToNormalizedUnitValue
}

// NormalizeUnitValue reports whether unit value normalization is enabled.
func (cc *CalculationConfig) NormalizeUnitValue() bool {
	return cc.normalizeUnitValue
}

// WithScaleToNormalizedUnitValue sets the target scale for unit value normalization.
func (cc *CalculationConfig) WithScaleToNormalizedUnitValue(scale int) {
	cc.scaleToNormalizedUnitValue = scale
}

// WithUnitValueNormalized enables unit value normalization.
func (cc *CalculationConfig) WithUnitValueNormalized() {
	cc.normalizeUnitValue = true
}

// WithNoUnitValueNormalization disables unit value normalization.
func (cc *CalculationConfig) WithNoUnitValueNormalization() {
	cc.normalizeUnitValue = false
}

// Input holds input values and tax definitions for a calculation.
type Input struct {
	UV      dec128.Dec128 // Unit Value
	GT      dec128.Dec128 // Gross Total
	QTY     dec128.Dec128 // Quantity
	Disc    dec128.Dec128 // Discount
	TaxList []*InputTax   // Taxes
}

// UnitValue returns the input unit value.
func (i *Input) UnitValue() dec128.Dec128 {
	return i.UV
}

// GrossTotal returns the input gross total.
func (i *Input) GrossTotal() dec128.Dec128 {
	return i.GT
}

// Qty returns the stored quantity.
func (i *Input) Qty() dec128.Dec128 {
	return i.QTY
}

// Discount returns the stored discount value.
func (i *Input) Discount() dec128.Dec128 {
	return i.Disc
}

// Taxes returns the input tax definitions.
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

// WithUnitValue updates the stored unit value value.
func (i *Input) WithUnitValue(uv dec128.Dec128) {
	i.UV = uv
}

// WithGrossTotal updates the stored gross total value.
func (i *Input) WithGrossTotal(gt dec128.Dec128) {
	i.GT = gt
}

// SetDiscToZero resets the input discount percentage to zero.
func (i *Input) SetDiscToZero() {
	i.Disc = dec128.FromInt(0)
}

// SetDiscToHundred sets the input discount percentage to one hundred.
func (i *Input) SetDiscToHundred() {
	i.Disc = dec128.FromInt(100)
}

// InputTax holds the tax data supplied with a calculation input.
type InputTax struct {
	CodeValue string        // Tax code
	NameValue string        // Tax name
	V         dec128.Dec128 // Tax value
	Id        int           // Tax ID
	Typee     Type          // Tax type
	Stagee    Stage         // Tax stage
}

// Stage implements TaxInformer.
func (it *InputTax) Stage() Stage {
	return it.Stagee
}

// Value implements TaxInformer.
func (it *InputTax) Value() dec128.Dec128 {
	return it.V
}

// ID returns the identifier of the tax or detail.
func (it *InputTax) ID() int {
	return it.Id
}

// Name returns the tax name.
func (it *InputTax) Name() string {
	return it.NameValue
}

// Code returns the tax code.
func (it *InputTax) Code() string {
	return it.CodeValue
}

// Type returns the tax calculation type.
func (it *InputTax) Type() Type {
	return it.Typee
}

// String returns the value as a string.
func (it *InputTax) String() string {
	js, _ := json.Marshal(it)
	return string(js)
}

var _ Enterable = (*Input)(nil)
var _ TaxInformer = (*InputTax)(nil)
