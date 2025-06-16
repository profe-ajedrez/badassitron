package withdec128

import "github.com/profe-ajedrez/badassitron/dec128"

type TaxStager interface {
	Validate(tx TaxInformer) error
	Bind(tx TaxInformer)
	Calc(taxable, qty dec128.Dec128) dec128.Dec128
}

// TaxInformer represn something that conatains information about a tax.
type TaxInformer interface {
	ID() int
	Name() string
	Code() string

	// Value returns the value of the tax.
	Value() dec128.Dec128

	// Type returns the type of the tax.
	// It can be Percentual, Amount, or AmountLine.
	// Percentual means the tax is a percentage of the unit value.
	// Amount means the tax is a fixed amount.
	// AmountLine means the tax is a fixed amount per line.
	// This is used to determine how the tax is applied.
	Type() Type

	// Stage() returns the stage of application of the tax.
	Stage() Stage

	String() string
}

type DetailTaxProcessor interface {
	Bind(qty dec128.Dec128, tx TaxInformer)
	Calc(taxableToInform, taxableToCalculate, qty dec128.Dec128)
	DetailTaxes() []TaxDetailer
}

type CalculationConfiger interface {
	Scale() int
	Flow() int
	NormalizeUnitValue() bool
	DetailTaxProcessor() DetailTaxProcessor
	WithDetailTaxProcessor(DetailTaxProcessor)
	WithUnitValueNormalized()
	WithNoUnitValueNormalization()
	UVNormalizer(dec128.Dec128) dec128.Dec128
	ValueNormalizer(dec128.Dec128) dec128.Dec128
}

type Enterable interface {
	UnitValue() dec128.Dec128
	GrossTotal() dec128.Dec128
	Qty() dec128.Dec128
	Discount() dec128.Dec128
	Taxes() []TaxInformer
	WithUnitValue(dec128.Dec128)
	WithGrossTotal(dec128.Dec128)

	SetDiscToZero()
	SetDiscToHundred()
}

type Outputable interface {
	Unitary() dec128.Dec128
	Qty() dec128.Dec128
	Net() dec128.Dec128
	Gross() dec128.Dec128
	Tax() dec128.Dec128
	Discount() dec128.Dec128
	GrossDiscount() dec128.Dec128
	DiscontedUnitary() dec128.Dec128
	NetWD() dec128.Dec128
	GrossWD() dec128.Dec128
	TaxWD() dec128.Dec128

	WithUnitary(dec128.Dec128)
	WithQty(dec128.Dec128)
	WithNet(dec128.Dec128)
	WithGross(dec128.Dec128)
	WithTax(dec128.Dec128)
	WithDiscount(dec128.Dec128)
	WithGrossDiscount(dec128.Dec128)
	WithDiscontedUnitary(dec128.Dec128)
	WithNetWD(dec128.Dec128)
	WithGrossWD(dec128.Dec128)
	WithTaxWD(dec128.Dec128)

	DetailTaxes() []TaxDetailer
	DetailDiscount() []DiscountDetailer

	WithTaxes([]TaxDetailer)
}

type TaxDetailer interface {
	Code() string
	Name() string
	RawAmount() dec128.Dec128
	Percent() dec128.Dec128
	Amount() dec128.Dec128
	Taxable() dec128.Dec128
	ID() int
	Type() Type

	WithCode(string)
	WithName(string)
	WithRawAmount(dec128.Dec128)
	WithPercent(dec128.Dec128)
	WithAmount(dec128.Dec128)
	WithTaxable(dec128.Dec128)
	WithID(int)
	WithType(Type)
}

type DiscountDetailer interface {
	Percent() dec128.Dec128
	Amount() dec128.Dec128
	RawPercent() dec128.Dec128
	Net() dec128.Dec128
	WithPercent(v dec128.Dec128)
	WithAmount(v dec128.Dec128)
	WithRawPercent(v dec128.Dec128)
	WithNet(v dec128.Dec128)
}

type HandlerFunc func(CalculationConfiger, Enterable, Outputable, ...HandlerFunc) error
