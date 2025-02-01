package badassitron

import (
	"github.com/alpacahq/alpacadecimal"
	"github.com/profe-ajedrez/badassitron/internal"
)

type TaxRater struct {
	next Stage
}

func NewTaxRater() *TaxRater {
	return &TaxRater{}
}

func (stage *TaxRater) SetNext(next Stage) {
	stage.next = next
}

func (stage *TaxRater) Execute(dt *Detail) (err error) {

	dt.TaxRatio, err = internal.Percentage(dt.Tax, dt.Net)

	if err != nil {
		return WrapWithWrappingError(err, " trying to obtain total tax rate ")
	}

	if stage.next != nil {
		return stage.next.Execute(dt)
	}

	return nil
}

// BruteUntaxer tries to remove the registered taxes from the brute value in the [Detail]
// storing it in the Net value of the [Detail]
//
// # Example
//
//	detail := Detail{
//	Uv: alpacadecimal.FromString("1000"),
//		Taxes: []TaxDetail{
//			{alpacadecimal.FromString("10"), Unit, alpacadecimal.Decimal0, alpacadecimal.Decimal0, 1, true},
//		},
//	}
//
//	untaxer := NewBruteUntaxer()
//	if err := untaxer.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("brute untaxed: ", detail.Net)
//
//	OUTPUT:
//	brute untaxed: 1111.1111111111111111111
type BruteUntaxer struct {
	next           Stage
	taxStageNumber int8
}

func NewBruteUntaxer(taxStageNumber int8) *BruteUntaxer {
	return &BruteUntaxer{nil, taxStageNumber}
}

func (d *BruteUntaxer) SetNext(next Stage) {
	d.next = next
}

func (d *BruteUntaxer) Execute(dt *Detail) (err error) {
	taxes := TaxesByStage(dt.Taxes, d.taxStageNumber)
	dt.Net = getTaxable(dt.Brute, dt.Qty, taxes)

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// UnitValueRounder handler able to round entry unitValue to the specified [detail.EntryUVScale]
type UnitValueRounder struct {
	next Stage
}

// NewUnitValueRounder returns a new UnitValueEntryProcessor stage able to round the unit value to the specified scale
func NewUnitValueRounder() *UnitValueRounder {
	return &UnitValueRounder{}
}

func (d *UnitValueRounder) SetNext(next Stage) {
	d.next = next
}

func (d *UnitValueRounder) Execute(dt *Detail) (err error) {

	if dt.EntryUVScale > emptyValue {
		dt.Uv = dt.Uv.Round(int32(dt.EntryUVScale))
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

type Unquantifier struct {
	next Stage
}

// NewUnquantifier returns a new Unquantifier stage able to get the unit value from netWd / qty
func NewUnquantifier() *Unquantifier {
	return &Unquantifier{}
}

func (d *Unquantifier) SetNext(next Stage) {
	d.next = next
}

func (d *Unquantifier) Execute(dt *Detail) (err error) {

	if dt.Qty.GreaterThan(alpacadecimal.Zero) {
		dt.Uv = dt.NetWd.Div(dt.Qty)
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// UnitValueUnDiscounter tries to remove the registered discounts from the unit value in the [Detail].
//
// # Example
//
//	detail := Detail{
//	Uv: alpacadecimal.FromString("1000"),
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("10"), true},
//		},
//	}
//
//	undiscounter := NewUnitValueUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("unit undiscounted: ", detail.UvWd)
//
//	OUTPUT:
//	net undiscounted: 1111.1111111111111111111
//
// Consider that you cant undiscount a unit value from a percentual discount of 100% without
// extra context, so you have to handle this case, may be in other [Stage] handler
//
// # Example discount 100%
//
//	detail := Detail{
//	Uv: alpacadecimal.FromString("0"), // we cannot know which was the original net without discount
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("100"), true},
//		},
//	}
//
//	undiscounter := NetUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("net undiscounted: ", detail.UvWd)
//
//	OUTPUT:
//	net undiscounted: 0
type UnitValueUnDiscounter struct {
	next Stage
}

// NewNetUnDiscounter returns a new UnDiscounter stage able to remove the discounts of a unit value
func NewUnitValueUnDiscounter() *UnitValueUnDiscounter {
	return &UnitValueUnDiscounter{}
}

func (d *UnitValueUnDiscounter) SetNext(next Stage) {
	d.next = next
}

func (d *UnitValueUnDiscounter) Execute(dt *Detail) (err error) {

	if dt.UvWd, err = discountable(dt.Uv, dt.Qty, dt.Discounts, "UnitValueUnDiscounter unit value: "); err != nil {

		return WrapWithWrappingError(err, "undiscounting unit value handler. "+dt.serialize())
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// NetUnDiscounter tries to remove the registered discounts from the net value in the [Detail].
//
// # Example
//
//	detail := Detail{
//	Net: alpacadecimal.FromString("1000"),
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("10"), true},
//		},
//	}
//
//	undiscounter := NewNetUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("net undiscounted: ", detail.NetWd)
//
//	OUTPUT:
//	net undiscounted: 1111.1111111111111111111
//
// Consider that you cant undiscount a net value from a percentual discount of 100% without
// extra context, so you have to handle this case, may be in other [Stage] handler
//
// # Example discount 100%
//
//	detail := Detail{
//	Net: alpacadecimal.FromString("0"), // we cannot know which was the original net without discount
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("100"), true},
//		},
//	}
//
//	undiscounter := NetUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("net undiscounted: ", detail.NetWd)
//
//	OUTPUT:
//	net undiscounted: 0
type NetUnDiscounter struct {
	next Stage
}

// NewNetUnDiscounter returns a new UnDiscounter stage able to remove the discounts of a net value
func NewNetUnDiscounter() *NetUnDiscounter {
	return &NetUnDiscounter{}
}

func (d *NetUnDiscounter) SetNext(next Stage) {
	d.next = next
}

func (d *NetUnDiscounter) Execute(dt *Detail) (err error) {

	if dt.Net.Equal(alpacadecimal.Zero) && len(dt.Discounts) > 0 {
		return WrapWithWrappingError(ErrCantUndiscountFromZero, " undiscounting net handler . "+dt.serialize())
	}

	if dt.NetWd, err = discountable(dt.Net, dt.Qty, dt.Discounts, "NetUndiscounter Net: "); err != nil {

		return WrapWithWrappingError(err, "undiscounting net handler. "+dt.serialize())
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// BruteUnDiscounter tries to remove the registered discounts from the brute value in the [Detail].
//
// # Example
//
//	detail := Detail{
//	Brute: alpacadecimal.FromString("1000"),
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("10"), true},
//		},
//	}
//
//	undiscounter := NewBruteUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("brute undiscounted: ", detail.BruteWd)
//
//	OUTPUT:
//	brute undiscounted: 1111.1111111111111111111
//
// Consider that you cant undiscount a brute value from a percentual discount of 100% without
// extra context, so you have to handle this case, may be in other [Stage] handler
//
// # Example discount 100%
//
//	detail := Detail{
//	Brute: alpacadecimal.FromString("0"), // we cannot know which was the original brute without discount
//		Discounts: []Discount{
//			{Unit, alpacadecimal.FromString("100"), true},
//		},
//	}
//
//	undiscounter := NewBruteUnDiscounter()
//	if err := undiscounter.Execute(&detail); err != nil {
//		panic(err) // dont panic!
//	}
//
//	fmt.Println("brute undiscounted: ", detail.BruteWd)
//
//	OUTPUT:
//	brute undiscounted: 0
type BruteUnDiscounter struct {
	next Stage
}

// NewDiscounter returns a new UnDiscounter stage able to remove the discounts of a value
func NewBruteUnDiscounter() *BruteUnDiscounter {
	return &BruteUnDiscounter{}
}

func (d *BruteUnDiscounter) SetNext(next Stage) {
	d.next = next
}

func (d *BruteUnDiscounter) Execute(dt *Detail) (err error) {

	if dt.BruteWd, err = discountable(dt.Brute, dt.Qty, dt.Discounts, "BruteUndiscounter Brute: "); err != nil {

		return WrapWithWrappingError(err, "undiscounting handler. "+dt.serialize())
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// Discounter is a handler able to calculate discounts and the net and netWd values
// Discounter applies the discounts over the netWd which is unitValue * quantity
// If no discounts are registered, only calculates net and makes netWd equal to net
type Discounter struct {
	next Stage
}

// NewDiscounter returns a new Discounter stage
func NewDiscounter() *Discounter {
	return &Discounter{}
}

func (d *Discounter) SetNext(next Stage) {
	d.next = next
}

func (d *Discounter) Execute(dt *Detail) error {
	dt.NetWd = dt.Uv.Mul(dt.Qty)

	amount, percentual, line := GroupDiscounts(dt.Discounts, dt.Qty)

	if err := DiscountBelowZero(amount, percentual, line); err != nil {
		return WrapWithWrappingError(err, "cummulated")
	}

	// protection against the case when percentual >= 100 avoid divition by zero
	dt.Net = alpacadecimal.Zero
	if percentual.LessThan(internal.Hundred) {
		part := internal.Hundred.Sub(percentual)
		dt.Net = internal.Part(dt.NetWd, part)
	}

	// protection against the case percentual + amount + line > 100% of value being sold
	dt.Net = dt.Net.Sub(amount).Sub(line)

	if dt.Net.LessThan(alpacadecimal.Zero) {
		dt.Net = alpacadecimal.Zero
	}

	dt.DiscountNetAmount = dt.NetWd.Sub(dt.Net)

	var err error
	dt.DiscountNetRatio, err = discountRatio(dt.Net, dt.NetWd, dt.DiscountNetAmount)

	if err != nil {
		return WrapWithWrappingError(err, "Discounter executing")
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

type BruteDiscounter struct {
	next Stage
}

// NewBruteDiscounter returns a new Discounter stage
func NewBruteDiscounter() *BruteDiscounter {
	return &BruteDiscounter{}
}

func (d *BruteDiscounter) SetNext(next Stage) {
	d.next = next
}

func (d *BruteDiscounter) Execute(dt *Detail) error {

	dt.DiscountAmount = dt.BruteWd.Sub(dt.Brute)

	var err error
	dt.DiscountRatio, err = discountRatio(dt.Brute, dt.BruteWd, dt.DiscountAmount)

	if err != nil {
		return WrapWithWrappingError(err, "Discounter executing")
	}

	if d.next != nil {
		return d.next.Execute(dt)
	}

	return nil
}

// TaxStage applies taxes You can apply how many stages you need
//
// ## Example:
//
// Perú has an amount tax named ICBP ehich apply over plastic bags, a tax called IGV which apply on every sale,
// and another called ISV which apply over some products.
// To calculate this taxes you can use a phormula like this:
//
// IVG(ISC(unit_value * qty) + unit_value * qty) + ISC(unit_value * qty) + ISBP
//
// which you can do in a serie of stages
//
// stage 1: isc = ISC(unit_value * qty)
// stage 2: igv = IGV(isc + unit_value * qty)
// stage 3: isbp = ISBP()
// stage 4: taxes = isc + igv + isbp
//
// So the application of taxes could be regulated by rules dependent of the customs of a country.
//
// To model this problem we could see the stages as a chain of responsability, exposing
// interfaces which define handlers behaviour, and implementing them to effectivelly perform the calculations
//
// Using [TaxStage] which implements the [Stage] interface, makes easy to model any tax configuration.
// In this example, a percentual tax applied in stage 2, and an amount tax applien in stage 1 over a taxable of 100
//
//		taxes := []TaxDetail{
//			{
//				Ratio: alpacadecimal.FromString("18"),
//				Applies: Unit,
//				Amount: alpacadecimal.Zero,
//				Taxable: alpacadecimal.Zero,
//				applyOn: 2,
//				Percentual: true,
//			},
//			{
//				Ratio: alpacadecimal.Zero,
//				Applies: Unit,
//				Amount: alpacadecimal.FromString("1"),
//				Taxable: alpacadecimal.Zero,
//				applyOn: 1,
//				Percentual: false,
//			},
//		}
//
//		unitValue := internal.Hundred
//		qty  := alpacadecimal.Decimal1
//
//		result := &Detail{ Uv: unitValue, Qty: qty, Taxes: taxes }
//
//		secondStage := NewTaxStage(2)    // define secondStage as a tax stage
//		firstStage := NewTaxStage(1)     // define firstStage as a tax stage
//		firstStage.SetNext(secondStage)  // register secondStage as the next step after completing the first
//
//		err := firstStage.Execute(result) // begin the chain of responsability
//		if err!=nil {
//			panic(err) // dont panic!
//		}
//
//	 js, _ := json.Marshal(result)
//		fmt.Println(string(js))
//
//		OUTPUT:
//	 { "unitValue": 100, "quantity": 1, "net": 100, "brute": 119.18, "tax": 19.18 }
type TaxStage struct {
	next        Stage
	stageNumber int8
}

func NewTaxStage(stNumber int8) *TaxStage {
	if stNumber < 0 {
		stNumber = 0
	}
	return &TaxStage{stageNumber: stNumber}
}

func (stage *TaxStage) SetNext(next Stage) {
	stage.next = next
}

func (stage *TaxStage) Execute(dt *Detail) error {

	commmontaxStageProcess(dt, stage.stageNumber)

	if stage.next != nil {
		return stage.next.Execute(dt)
	}

	return nil
}

// BruteSnapshot is a handler able to calculate the brute and brute without discount values.
// should be called ater discounts and taxes are applied or else these values could be wrong
// Normally you'd call for it at the end of the chain of responsability
type BruteSnapshot struct {
	next Stage
}

// NewBruteSnapshot calculates the brute and brute without discount subtotals
// from the net, netWd, tax and taxWd values
func NewBruteSnapshot() *BruteSnapshot {
	return &BruteSnapshot{}
}

func (stage *BruteSnapshot) SetNext(next Stage) {
	stage.next = next
}

func (stage *BruteSnapshot) Execute(dt *Detail) error {

	dt.Brute = dt.Net.Add(dt.Tax)
	dt.BruteWd = dt.NetWd.Add(dt.TaxWd)

	if stage.next != nil {
		return stage.next.Execute(dt)
	}

	return nil
}

// commmontaxStageProcess does the common logic of a tax stage
//
// 1 obtain the taxes of the current stage
//
// 2 get the taxable of the current stage. generally speaking, net + sum|previous_stages.taxes|
//
// 3 set the values of taxes in the taxes detail
//
// 4 obtain the values of amount, percentual and line taxes
//
// 5 add this values to the current value of tax
//
// 6 The same for taxWd
func commmontaxStageProcess(dt *Detail, stage int8) {
	// stageredTaxes are the taxes associated to this stage
	stageredTaxes := TaxesByStage(dt.Taxes, stage)
	// stageTaxable is the taxable of this stage
	stageTaxable := dt.Net.Add(dt.Tax)

	SetTaxesValues(stageredTaxes, stageTaxable)
	amount, perc, line := GroupTaxes(stageredTaxes, dt.Qty)

	// tax is the addition of the previous stages taxes with the current stage taxes
	// tax = tax + amount + line + (stageTaxable * perc / 100)
	dt.Tax = dt.Tax.Add(amount).Add(line).Add(internal.Part(stageTaxable, perc))
	// taxWd is the addition of the previous stages taxes without discounts with the current stage taxes without discounts
	dt.TaxWd = dt.TaxWd.Add(amount).Add(line).Add(internal.Part(dt.NetWd.Add(dt.TaxWd), perc))

	k := 0
	for j := range dt.Taxes {
		if dt.Taxes[j].Stage == stage {
			dt.Taxes[j].Amount = stageredTaxes[k].Amount
			dt.Taxes[j].Percentual = stageredTaxes[k].Percentual
			dt.Taxes[j].Ratio = stageredTaxes[k].Ratio
			dt.Taxes[j].Taxable = stageredTaxes[k].Taxable
			dt.Taxes[j].Applies = stageredTaxes[k].Applies
			k += 1
		}
	}

}

func discountRatio(net, netWd, discountAmount alpacadecimal.Decimal) (alpacadecimal.Decimal, error) {
	discountRatio := alpacadecimal.Zero

	if discountAmount.GreaterThan(alpacadecimal.Zero) {
		var err error
		discountRatio, err = internal.Percentage(discountAmount, netWd)

		if err != nil {
			return discountRatio, WrapWithWrappingError(err, "calculating total tiscount ratio")
		}
	} else if net.Equal(alpacadecimal.Zero) && netWd.GreaterThan(alpacadecimal.Zero) {
		discountRatio = internal.Hundred
	}
	return discountRatio, nil
}

func discountable(value, qty alpacadecimal.Decimal, discounts []Discount, msg string) (alpacadecimal.Decimal, error) {
	amount, percentual, line := GroupDiscounts(discounts, qty)

	part := value.Add(amount).Add(line)
	perc := internal.Hundred.Sub(percentual)

	if perc.Equal(alpacadecimal.Zero) {
		return part, nil
	}

	discountable, err := internal.Total(part, perc)

	if err != nil {
		sb := internal.GetSB()
		defer internal.PutSB(sb)

		sb.WriteString("undiscounting perc: ")
		sb.WriteString(perc.String())
		sb.WriteString(" from ")
		sb.WriteString(msg)
		sb.WriteString(" ")
		sb.WriteString(value.String())
		return alpacadecimal.Zero, WrapWithWrappingError(err, sb.String())
	}

	return discountable, nil
}

func getTaxable(value, qty alpacadecimal.Decimal, taxes []TaxDetail) alpacadecimal.Decimal {

	amount, percentual, line := GroupTaxes(taxes, qty)

	part := value.Sub(amount).Sub(line)
	divisor := internal.One.Add(percentual.Div(internal.Hundred))

	taxable := part.Div(divisor)

	return taxable
}
