package forwarder

func (p *Proxy) MustRun() {
	if err := p.Run(); err != nil {
		panic(err)
	}
}
