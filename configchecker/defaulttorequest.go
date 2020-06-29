package configchecker

func (c *ConfigChecker) defaultToRequest() []*configOption {
	setColors()

	return []*configOption{
		{
			Tag:           "#DATABASE_HOST",
			Name:          "Database host",
			Info:          lightBlueString("0.0.0.0") + ", " + lightBlueString("127.0.0.1") + ", " + darkBlueString("localhost"),
			DefaultOption: "localhost",
		},
		{
			Tag:           "#DATABASE_PORT",
			Name:          "Database port",
			Info:          darkBlueString("26257"),
			DefaultOption: "26257",
		},
		{
			Tag:           "#DATABASE_NAME",
			Name:          "Database name",
			Info:          darkBlueString("app"),
			DefaultOption: "app",
		},
		{
			Tag:  "#EMAIL_HOST",
			Name: "Email host",
			Info: "smtp.domain.com",
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
			Info: "noreply@domain.com",
		},
	}
}
