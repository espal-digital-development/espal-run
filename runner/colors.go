package runner

func (r *Runner) loadColors() {
	r.logColors = map[string]string{
		"reset":          "0",
		"black":          "30",
		"red":            "31",
		"green":          "32",
		"yellow":         "33",
		"blue":           "34",
		"magenta":        "35",
		"cyan":           "36",
		"white":          "37",
		"bold_black":     "30;1",
		"bold_red":       "31;1",
		"bold_green":     "32;1",
		"bold_yellow":    "33;1",
		"bold_blue":      "34;1",
		"bold_magenta":   "35;1",
		"bold_cyan":      "36;1",
		"bold_white":     "37;1",
		"bright_black":   "30;2",
		"bright_red":     "31;2",
		"bright_green":   "32;2",
		"bright_yellow":  "33;2",
		"bright_blue":    "34;2",
		"bright_magenta": "35;2",
		"bright_cyan":    "36;2",
		"bright_white":   "37;2",
	}
}
