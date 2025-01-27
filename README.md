# Badassitron

<br />
<div align="center">
  <a href="./internal/badassitron-logo.png">
    <img src="./internal/badassitron-logo.png" alt="badassitron Logo" width="300" height="auto">
  </a>


![](internal/optimus_prime_says_his_line%20.png)

The more than meets the eye, no nonsense chain of responsability based library to implement modern sales calculation 'til all are one.

And yes, the name is for that [scene](https://youtu.be/5a09yJU-mCI?si=YbvwWdjOpnWg6IIH&t=42) in the Transformers One movie.


But, [What is chain of responsability about?](https://www.geeksforgeeks.org/chain-responsibility-design-pattern/)


But, this look like middlewares? Yes! that is it! i'ts the same pattern.

## Motivation

We see a lot of legacy sales system at my job. In them, the calculations of important values of the process are convoluted and coupled with the data model, which makes it very difficult and laborious to modify parts of it, adapt it or check for bugs.

Seeing this over a long period of time, we began to think about how to organize the process of these calculations so that it was unintrusive, decoupled and possible to dynamically compose according to different requirements.

So that's where our motivation comes from for this little side project.


## Installation


Install in your project with go get

```bash
$ go get github.com/profe-ajedrez/badassitron
```

## Usage


You can use some of the included stages to cosntruct your calculation flow, or code your own. You just have to make them implement the stage [Stage] interface.


### Example

You want to calculate the subtotals from an unit value, quantity, some discounts and taxes.

We just have to invoke the handlers needed for the flow we are implementing and pass the detail at the first of them

```go
    detail :=  Detail{
        Uv:  func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(),
        Qty: alpacadecimal.NewFromFloat(2.5),
        Discounts: []Discount{
            {Unit, func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("10"); return d }(), true},
        },
        Taxes: []TaxDetail{
            {func() alpacadecimal.Decimal { d, _ := alpacadecimal.NewFromString("19"); return d }(), Unit, alpacadecimal.Zero, alpacadecimal.Zero, 2, true},					
        },
        EntryUVScale: 2,
    }

    // Defining a third and last stage to calculate taxes (taxes that are not overtaxes but should be calculated after them)
    third := NewTaxStage(3)
    third.SetNext()

    // Defining a second stage to calculate taxes (overtaxes)
    second := NewTaxStage(2)
    second.SetNext(third)

    // Defining a first stage for taxes
    first := NewTaxStage(1)
    first.SetNext(second)

    // Defining a stage to calculate discounts
    discHandler := NewDiscounter()
    discHandler.SetNext(first)

    // Did you note that discHandler was the last handler defined, but the first called?
    err := discHandler.Execute(&detail)

```

