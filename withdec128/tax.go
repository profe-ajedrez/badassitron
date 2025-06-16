package withdec128

import (
	"github.com/profe-ajedrez/badassitron/dec128"
)

type Stages struct {
	Natural NaturalTaxStage
	Overtax OverTaxStage
	Bypass  BypassTaxStage
	Invalid InvalidStage
}

func NewTaxStages() *Stages {
	return &Stages{
		Natural: NaturalTaxStage{&TaxStage{}},
		Overtax: OverTaxStage{&TaxStage{}},
		Bypass:  BypassTaxStage{&TaxStage{}},
		Invalid: InvalidStage{&TaxStage{}},
	}
}

func (s *Stages) Bind(qty dec128.Dec128, tx TaxInformer) error {
	if tx == nil {
		return NewTaxError(ErrNilArgument, "la información recibida de impuesto es nil")
	}

	switch tx.Stage() {
	case Natural:
		if err := s.Natural.Validate(tx); err != nil {
			return err
		}
		s.Natural.Bind(qty, tx)
		return nil
	case Overtax:
		if err := s.Overtax.Validate(tx); err != nil {
			return err
		}
		s.Overtax.Bind(qty, tx)
	case Bypass:
		if err := s.Bypass.Validate(tx); err != nil {
			return err
		}
		s.Bypass.Bind(qty, tx)
	default:
		s.Invalid.Bind(qty, tx)
		return s.Invalid.Validate(tx)
	}

	return nil
}

type InvalidStage struct {
	*TaxStage
}

func (n *InvalidStage) Validate(tx TaxInformer) error {
	return NewTaxError(ErrInvalidTaxStage, "el stage del impuesto no es válido: "+tx.String())
}

type NaturalTaxStage struct {
	*TaxStage
}

func (n *NaturalTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewNaturalTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

type OverTaxStage struct {
	*TaxStage
}

func (n *OverTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewOverTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

type BypassTaxStage struct {
	*TaxStage
}

func (n *BypassTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewBypassTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

type TaxStage struct {
	amount  dec128.Dec128
	percent dec128.Dec128
}

func (n *TaxStage) Validate(tx TaxInformer) error {
	if tx == nil {
		return NewTaxError(ErrNilArgument, "la información recibida de impuesto es nil")
	}

	if tx.Value().IsNegative() {
		return ErrNegativeTax
	}

	if tx.Type() != Percentual && tx.Type() != Amount && tx.Type() != AmountLine {
		return ErrInvalidTaxType
	}

	if tx.Stage() < 0 || tx.Stage() > 2 {
		return ErrTaxStageOutOfBounds
	}

	if tx.Type() == Percentual && tx.Value().GreaterThan(Hundred()) {
		return ErrTaxOver100
	}

	return nil
}

func (t *TaxStage) Bind(qty dec128.Dec128, tx TaxInformer) {
	if tx.Type() == Percentual {
		t.percent = t.percent.Add(tx.Value())
	}

	if tx.Type() == Amount {
		t.amount = t.amount.Add(tx.Value().Mul(qty))
	}

	if tx.Type() == AmountLine {
		t.amount = t.amount.Add(tx.Value())
	}
}

func (t *TaxStage) Calc(taxable, qty dec128.Dec128) dec128.Dec128 {
	r := t.percent.Div(dec128.Decimal100)
	return taxable.Mul(r).Add(t.amount)
}

type DetailTaxes struct {
	list map[int]TaxDetailer
}

func (dt *DetailTaxes) DetailTaxes() []TaxDetailer {
	details := make([]TaxDetailer, 0, len(dt.list))
	for _, tax := range dt.list {
		details = append(details, tax)
	}
	return details
}

func NewDetailTaxes() *DetailTaxes {
	return &DetailTaxes{
		list: make(map[int]TaxDetailer),
	}
}

func (dt *DetailTaxes) Bind(qty dec128.Dec128, tx TaxInformer) {

	dt.list[tx.ID()] = &DetailTax{
		code:      tx.Code(),
		name:      tx.Name(),
		rawAmount: Zero(),
		percent:   Zero(),
		amount:    Zero(),
		id:        tx.ID(),
		typee:     tx.Type(),
	}

	if tx.Type() == Percentual {
		dt.list[tx.ID()].WithPercent(tx.Value())
	}

	if tx.Type() == Amount {
		dt.list[tx.ID()].WithAmount(tx.Value().Mul(qty))
	}

	if tx.Type() == AmountLine {
		dt.list[tx.ID()].WithAmount(tx.Value())
	}
}

func (dt *DetailTaxes) Calc(taxableToInform, taxableToCalculate, qty dec128.Dec128) {
	for _, tax := range dt.list {
		if tax.Type() == Percentual {
			porcentualAmount := taxableToCalculate.Mul(tax.Percent().Div(dec128.Decimal100)).Mul(qty)
			tax.WithRawAmount(porcentualAmount)
			tax.WithAmount(porcentualAmount)
		} else if tax.Type() == Amount {
			ratio := tax.Amount().Mul(dec128.Decimal100).Div(taxableToCalculate)
			tax.WithPercent(ratio)
		}
		tax.WithTaxable(taxableToInform)
	}
}

type DetailTax struct {
	code      string
	name      string
	taxable   dec128.Dec128
	rawAmount dec128.Dec128
	percent   dec128.Dec128
	amount    dec128.Dec128
	id        int
	typee     Type
}

func (dt *DetailTax) Code() string {
	return dt.code
}

func (dt *DetailTax) Name() string {
	return dt.name
}

func (dt *DetailTax) Taxable() dec128.Dec128 {
	return dt.taxable
}

func (dt *DetailTax) RawAmount() dec128.Dec128 {
	return dt.rawAmount
}

func (dt *DetailTax) Percent() dec128.Dec128 {
	return dt.percent
}

func (dt *DetailTax) Amount() dec128.Dec128 {
	return dt.amount
}

func (dt *DetailTax) ID() int {
	return dt.id
}

func (dt *DetailTax) Type() Type {
	return dt.typee
}

func (dt *DetailTax) WithCode(code string) {
	dt.code = code
}

func (dt *DetailTax) WithName(nm string) {
	dt.name = nm
}

func (dt *DetailTax) WithRawAmount(raw dec128.Dec128) {
	dt.rawAmount = raw
}

func (dt *DetailTax) WithTaxable(v dec128.Dec128) {
	dt.taxable = v
}

func (dt *DetailTax) WithPercent(p dec128.Dec128) {
	dt.percent = p
}

func (dt *DetailTax) WithAmount(a dec128.Dec128) {
	dt.amount = a
}

func (dt *DetailTax) WithID(id int) {
	dt.id = id
}

func (dt *DetailTax) WithType(tp Type) {
	dt.typee = tp
}

var _ TaxDetailer = &DetailTax{}
var _ DetailTaxProcessor = &DetailTaxes{}
