package withfloat64

// Stages groups the available tax stages for a calculation.
type Stages struct {
	Natural NaturalTaxStage
	Overtax OverTaxStage
	Bypass  BypassTaxStage
	Invalid InvalidStage
}

// NewTaxStages returns a fresh set of tax stage accumulators.
func NewTaxStages() *Stages {
	return &Stages{
		Natural: NaturalTaxStage{&TaxStage{}},
		Overtax: OverTaxStage{&TaxStage{}},
		Bypass:  BypassTaxStage{&TaxStage{}},
		Invalid: InvalidStage{&TaxStage{}},
	}
}

// Bind adds a tax to the appropriate stage or tax detail collection.
func (s *Stages) Bind(qty float64, tx TaxInformer) error {
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

// InvalidStage handles tax definitions with an unsupported stage.
type InvalidStage struct {
	*TaxStage
}

// Validate checks that a tax definition meets the stage requirements.
func (n *InvalidStage) Validate(tx TaxInformer) error {
	return NewTaxError(ErrInvalidTaxStage, "el stage del impuesto no es válido: "+tx.String())
}

// NaturalTaxStage handles taxes applied to the initial taxable amount.
type NaturalTaxStage struct {
	*TaxStage
}

// Validate checks that a tax definition meets the stage requirements.
func (n *NaturalTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewNaturalTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

// OverTaxStage handles taxes applied after natural taxes.
type OverTaxStage struct {
	*TaxStage
}

// Validate checks that a tax definition meets the stage requirements.
func (n *OverTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewOverTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

// BypassTaxStage handles taxes applied to the base independently of natural taxes.
type BypassTaxStage struct {
	*TaxStage
}

// Validate checks that a tax definition meets the stage requirements.
func (n *BypassTaxStage) Validate(tx TaxInformer) error {
	if err := n.TaxStage.Validate(tx); err != nil {
		return NewBypassTaxError(err, "error en impuesto natural: "+tx.String())
	}
	return nil
}

// TaxStage accumulates percentage and fixed tax amounts for one stage.
type TaxStage struct {
	amount  float64
	percent float64
}

// Validate checks that a tax definition meets the stage requirements.
func (n *TaxStage) Validate(tx TaxInformer) error {
	if tx == nil {
		return NewTaxError(ErrNilArgument, "la información recibida de impuesto es nil")
	}

	if tx.Value() < 0 {
		return ErrNegativeTax
	}

	if tx.Type() != Percentual && tx.Type() != Amount && tx.Type() != AmountLine {
		return ErrInvalidTaxType
	}

	if tx.Stage() < 0 || tx.Stage() > 2 {
		return ErrTaxStageOutOfBounds
	}

	if tx.Type() == Percentual && tx.Value() > 100 {
		return ErrTaxOver100
	}

	return nil
}

// Bind adds a tax to the appropriate stage or tax detail collection.
func (t *TaxStage) Bind(qty float64, tx TaxInformer) {
	if tx.Type() == Percentual {
		t.percent = t.percent + tx.Value()
	}

	if tx.Type() == Amount {
		t.amount = t.amount + tx.Value()*qty
	}

	if tx.Type() == AmountLine {
		t.amount = t.amount + tx.Value()
	}
}

// Calc calculates the tax or tax detail values using the current state.
func (t *TaxStage) Calc(taxable, qty float64) float64 {
	r := t.percent / 100
	return taxable*r + t.amount
}

// DetailTaxes holds the calculated details for individual taxes.
type DetailTaxes struct {
	list map[int]TaxDetailer
}

// DetailTaxes returns the recorded individual tax details.
func (dt *DetailTaxes) DetailTaxes() []TaxDetailer {
	details := make([]TaxDetailer, 0, len(dt.list))
	for _, tax := range dt.list {
		details = append(details, tax)
	}
	return details
}

// NewDetailTaxes returns an empty collection of itemized tax details.
func NewDetailTaxes() *DetailTaxes {
	return &DetailTaxes{
		list: make(map[int]TaxDetailer),
	}
}

// Bind adds a tax to the appropriate stage or tax detail collection.
func (dt *DetailTaxes) Bind(qty float64, tx TaxInformer) {

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
		dt.list[tx.ID()].WithAmount(tx.Value() * qty)
	}

	if tx.Type() == AmountLine {
		dt.list[tx.ID()].WithAmount(tx.Value())
	}
}

// Calc calculates the tax or tax detail values using the current state.
func (dt *DetailTaxes) Calc(taxableToInform, taxableToCalculate, qty float64) {
	for _, tax := range dt.list {
		if tax.Type() == Percentual {
			porcentualAmount := taxableToCalculate * tax.Percent() / 100 * qty
			tax.WithRawAmount(porcentualAmount)
			tax.WithAmount(porcentualAmount)
		} else if tax.Type() == Amount {
			ratio := tax.Amount() * 100 / taxableToCalculate
			tax.WithPercent(ratio)
		}
		tax.WithTaxable(taxableToInform)
	}
}

// DetailTax stores the calculated values and metadata of one tax.
type DetailTax struct {
	code      string
	name      string
	taxable   float64
	rawAmount float64
	percent   float64
	amount    float64
	id        int
	typee     Type
}

// Code returns the tax code.
func (dt *DetailTax) Code() string {
	return dt.code
}

// Name returns the tax name.
func (dt *DetailTax) Name() string {
	return dt.name
}

// Taxable returns the recorded taxable base.
func (dt *DetailTax) Taxable() float64 {
	return dt.taxable
}

// RawAmount returns the unadjusted tax amount.
func (dt *DetailTax) RawAmount() float64 {
	return dt.rawAmount
}

// Percent returns the stored percentage.
func (dt *DetailTax) Percent() float64 {
	return dt.percent
}

// Amount returns the stored amount.
func (dt *DetailTax) Amount() float64 {
	return dt.amount
}

// ID returns the identifier of the tax or detail.
func (dt *DetailTax) ID() int {
	return dt.id
}

// Type returns the tax calculation type.
func (dt *DetailTax) Type() Type {
	return dt.typee
}

// WithCode updates the stored code value.
func (dt *DetailTax) WithCode(code string) {
	dt.code = code
}

// WithName updates the stored name value.
func (dt *DetailTax) WithName(nm string) {
	dt.name = nm
}

// WithRawAmount updates the stored raw amount value.
func (dt *DetailTax) WithRawAmount(raw float64) {
	dt.rawAmount = raw
}

// WithTaxable updates the stored taxable value.
func (dt *DetailTax) WithTaxable(v float64) {
	dt.taxable = v
}

// WithPercent updates the stored percent value.
func (dt *DetailTax) WithPercent(p float64) {
	dt.percent = p
}

// WithAmount updates the stored amount value.
func (dt *DetailTax) WithAmount(a float64) {
	dt.amount = a
}

// WithID updates the stored id value.
func (dt *DetailTax) WithID(id int) {
	dt.id = id
}

// WithType updates the stored type value.
func (dt *DetailTax) WithType(tp Type) {
	dt.typee = tp
}

var _ TaxDetailer = &DetailTax{}
var _ DetailTaxProcessor = &DetailTaxes{}
