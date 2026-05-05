module github.com/verzac/grocer-discord-bot

go 1.25.6

require (
	github.com/andanhm/go-prettytime v1.1.0
	github.com/aws/aws-sdk-go v1.40.32
	github.com/bwmarrin/discordgo v0.29.0
	github.com/go-playground/validator/v10 v10.10.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/joho/godotenv v1.3.0
	github.com/labstack/echo-contrib v0.50.1
	github.com/labstack/echo/v4 v4.15.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/stretchr/testify v1.11.1
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.47.0
	golang.org/x/oauth2 v0.34.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.14
)

// note: see decision-docs/001-pin-discordgo-to-v0.29.0.md for more details on why this is needed
replace github.com/bwmarrin/discordgo => github.com/verzac/discordgo-new-components v0.29.1-0.20260505070713-be8e18feaa36

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/google/go-github/v35 v35.2.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/cc/v3 v3.32.4 // indirect
	modernc.org/ccgo/v3 v3.9.2 // indirect
	modernc.org/libc v1.9.5 // indirect
	modernc.org/mathutil v1.2.2 // indirect
	modernc.org/memory v1.0.4 // indirect
	modernc.org/opt v0.1.1 // indirect
	modernc.org/sqlite v1.10.6 // indirect
	modernc.org/strutil v1.1.0 // indirect
	modernc.org/token v1.0.0 // indirect
)
