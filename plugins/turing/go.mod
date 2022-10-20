module github.com/caraml-dev/xp/plugins/turing

go 1.18

require (
	bou.ke/monkey v1.0.2
	github.com/caraml-dev/turing/engines/experiment v0.0.0-20220816023749-cf36433c873b
	github.com/caraml-dev/xp/clients v0.0.0-00010101000000-000000000000
	github.com/caraml-dev/xp/common v0.0.0
	github.com/caraml-dev/xp/treatment-service v0.0.0
	github.com/go-playground/validator/v10 v10.11.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	go.uber.org/zap v1.19.1
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.11.0 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/getkin/kin-openapi v0.94.0 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-chi/chi/v5 v5.0.7 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible // indirect
	github.com/gojek/mlp v1.5.3 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-hclog v0.16.0 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/newrelic/go-agent v3.9.0+incompatible // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/zaffka/zap-to-hclog v0.10.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220513210258-46612604a0f9 // indirect
	golang.org/x/net v0.0.0-20220513224357-95641704303c // indirect
	golang.org/x/sys v0.0.0-20220513210249-45d2b4557a2a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3 // indirect
	google.golang.org/grpc v1.46.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	github.com/caraml-dev/xp/clients => ../../clients
	github.com/caraml-dev/xp/common => ../../common
	github.com/caraml-dev/xp/treatment-service => ../../treatment-service
)
