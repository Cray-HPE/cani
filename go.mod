module github.com/Cray-HPE/cani

go 1.20

replace github.com/Cray-HPE/cani/pkg/sls-client => ./pkg/sls-client

replace github.com/Cray-HPE/cani/pkg/hsm-client => ./pkg/hsm-client

replace github.com/Cray-HPE/cani/pkg/sls-plugin => ./pkg/sls-plugin

replace internal/shell => ./internal/shell

require internal/hsm v1.0.0

replace internal/hsm => ./internal/hsm

require internal/sls v1.0.0

replace internal/sls => ./internal/sls

require internal/shellspec v1.0.0

replace internal/shellspec => ./internal/shellspec

require (
	github.com/Cray-HPE/hms-xname v1.1.0
	github.com/antihax/optional v1.0.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/rs/zerolog v1.29.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.0
	github.com/spf13/cobra v1.7.0
	golang.org/x/oauth2 v0.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
