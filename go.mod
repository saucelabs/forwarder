module github.com/saucelabs/forwarder

go 1.19

require (
	github.com/dop251/goja v0.0.0-20221118162653-d4bf6fde1b86
	github.com/google/go-cmp v0.5.9
	github.com/google/martian/v3 v3.1.0
	github.com/gorilla/websocket v1.5.0
	github.com/kevinburke/hostsfile v0.0.0-20220522040509-e5e984885321
	github.com/mmatczuk/anyflag v0.0.0-20221011091730-9f8baeb6fd9c
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/common v0.37.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0
	go.uber.org/atomic v1.10.0
	go.uber.org/goleak v1.2.0
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/google/martian/v3 => github.com/saucelabs/martian/v3 v3.0.0-20221206110303-9457eff1a5fd
