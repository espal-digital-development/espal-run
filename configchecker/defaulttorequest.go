package configchecker

func (c *ConfigChecker) defaultToRequest() []*configOption {
	return []*configOption{
		{
			Tag:  "#DATABASE_HOST",
			Name: "Database host",
			Info: c.textPainter.lightBlueString("0.0.0.0") + ", " +
				c.textPainter.lightBlueString("127.0.0.1") + ", " + c.textPainter.darkBlueString("localhost"),
			DefaultOption: "localhost",
		},
		{
			Tag:           "#DATABASE_PORT",
			Name:          "Database port",
			Info:          c.textPainter.darkBlueString("36257"),
			DefaultOption: "36257",
		},
		{
			Tag:           "#DATABASE_NAME",
			Name:          "Database name",
			Info:          c.textPainter.darkBlueString("app"),
			DefaultOption: "app",
		},
		{
			Tag:  "#EMAIL_HOST",
			Name: "Email host",
			Info: "smtp.domain.dev",
		},
		{
			Tag:           "#EMAIL_PORT",
			Name:          "Email port",
			DefaultOption: "2525",
		},
		{
			Tag:  "#EMAIL_USERNAME",
			Name: "Email username",
		},
		{
			Tag:  "#EMAIL_PASSWORD",
			Name: "Email password",
		},
		{
			Tag:  "#EMAIL_NO_REPLY_ADDRESS",
			Name: "Email no-reply address",
			Info: "noreply@domain.dev",
		},
	}
}
