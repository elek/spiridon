package satellite

type Config struct {
	TelegramToken string `koanf:"telegram_token"`
	DatabaseDir   string `koanf:"database_dir"`
	WebPort       int    `koanf:"web_port"`
	DrpcPort      int    `koanf:"drpc_port"`
	Db            string `koanf:"db"`
	CookieSecret  string `koanf:"cookie_secret"`
	Domain        string `koanf:"domain"`
}
