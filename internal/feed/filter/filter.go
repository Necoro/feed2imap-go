package filter

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/mmcdole/gofeed"
)

type Filter struct {
	prog *vm.Program
}

func (f *Filter) Run(item *gofeed.Item) (bool, error) {
	if res, err := expr.Run(f.prog, item); err != nil {
		return false, err
	} else {
		return res.(bool), nil
	}
}

func New(s string) (*Filter, error) {
	prog, err := expr.Compile(s, expr.AsBool(), expr.Env(gofeed.Item{}))
	if err != nil {
		return nil, err
	}
	return &Filter{prog}, nil
}
