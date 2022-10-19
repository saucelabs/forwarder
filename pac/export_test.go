package pac

import "github.com/dop251/goja"

func (pr *ProxyResolver) TestingEval(script string) (goja.Value, error) {
	return pr.vm.RunString(script)
}
