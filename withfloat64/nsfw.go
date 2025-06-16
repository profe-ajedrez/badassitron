package withfloat64

type TaxStager interface {
	Validate(tx TaxInformer) error
	Bind(tx TaxInformer)
	Calc(taxable, qty float64) float64
}

// TaxInformer represn something that conatains information about a tax.
type TaxInformer interface {
	ID() int
	Name() string
	Code() string

	// Value returns the value of the tax.
	Value() float64

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
	Bind(qty float64, tx TaxInformer)
	Calc(taxableToInform, taxableToCalculate, qty float64)
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
	UVNormalizer(float64) float64
	ValueNormalizer(float64) float64
}

type Enterable interface {
	UnitValue() float64
	GrossTotal() float64
	Qty() float64
	Discount() float64
	Taxes() []TaxInformer
	WithUnitValue(float64)
	WithGrossTotal(float64)

	SetDiscToZero()
	SetDiscToHundred()
}

type Outputable interface {
	Unitary() float64
	Qty() float64
	Net() float64
	Gross() float64
	Tax() float64
	Discount() float64
	GrossDiscount() float64
	DiscontedUnitary() float64
	NetWD() float64
	GrossWD() float64
	TaxWD() float64

	WithUnitary(float64)
	WithQty(float64)
	WithNet(float64)
	WithGross(float64)
	WithTax(float64)
	WithDiscount(float64)
	WithGrossDiscount(float64)
	WithDiscontedUnitary(float64)
	WithNetWD(float64)
	WithGrossWD(float64)
	WithTaxWD(float64)

	DetailTaxes() []TaxDetailer
	DetailDiscount() []DiscountDetailer

	WithTaxes([]TaxDetailer)
}

type TaxDetailer interface {
	Code() string
	Name() string
	RawAmount() float64
	Percent() float64
	Amount() float64
	Taxable() float64
	ID() int
	Type() Type

	WithCode(string)
	WithName(string)
	WithRawAmount(float64)
	WithPercent(float64)
	WithAmount(float64)
	WithTaxable(float64)
	WithID(int)
	WithType(Type)
}

type DiscountDetailer interface {
	Percent() float64
	Amount() float64
	RawPercent() float64
	Net() float64
	WithPercent(v float64)
	WithAmount(v float64)
	WithRawPercent(v float64)
	WithNet(v float64)
}

type HandlerFunc func(CalculationConfiger, Enterable, Outputable, ...HandlerFunc) error
