// Package withdec128 provides the existing staged calculation workflow built
// on Dec128. Callers supply calculation input, output, and configuration
// through interfaces, then run handlers in a chosen order.
//
// This is a separate workflow package. The numeric decimal packages do not
// contain tax or discount policy. Some older API paths in this package are
// incomplete. Consult the calculation workflow guide before using them.
package withdec128
