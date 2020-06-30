package configchecker

type textPainter struct {
	reset     string
	lightBlue string
	darkBlue  string
}

func (p *textPainter) resolveDefaults() {
}

func (p *textPainter) lightBlueString(value string) string {
	return p.lightBlue + value + p.reset
}

func (p *textPainter) darkBlueString(value string) string {
	return p.darkBlue + value + p.reset
}
