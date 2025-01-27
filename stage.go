package badassitron

// Stage represents a handler able to make some kind of calculation over a pointer to [Detail]
// Stage can set the next handler to be called when its job is done.
//
// Stage let us implement a chain of responsability
type Stage interface {
	// Execute performs the work over the [Detail]
	Execute(*Detail) error

	// SetNext sets the next Pricer to be called after the current work is done
	SetNext(Stage)
}
