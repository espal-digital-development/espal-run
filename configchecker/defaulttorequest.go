package configchecker

func (c *ConfigChecker) defaultToRequest() []*configOption {
	return []*configOption{
		{
			Tag:           "#DATABASE_HOST",
			Name:          "Database host",
			Info:          "\033[0;34m0.0.0.0\033[m, \033[0;34m127.0.0.1\033[m, \033[0;94mlocalhost\033[m",
			DefaultOption: "localhost",
		},
		{
			Tag:           "#DATABASE_PORT",
			Name:          "Database port",
			Info:          "\033[0;94m26257\033[m",
			DefaultOption: "26257",
		},
		{
			Tag:           "#DATABASE_NAME",
			Name:          "Database name",
			Info:          "\033[0;94mapp\033[m",
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
