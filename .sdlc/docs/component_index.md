# Component Index

main (cmd/api)
├── Config [struct] (config.go)
│   └── Validate (Validate() error)
├── LoadConfig [func] (config.go)
├── TestConfig_Validate [func] (config_test.go)
├── TestLoadConfig_Defaults [func] (config_test.go)
├── TestLoadConfig_MissingDatabaseURL [func] (config_test.go)
├── TestLoadConfig_PortFromEnv [func] (config_test.go)
└── main [func] (main.go)

platform (internal/platform)
├── NewPool [func] (db.go)
├── TestNewPool_ContextCancelled [func] (db_test.go)
├── TestNewPool_InvalidDSN [func] (db_test.go)
├── TestNewPool_ReachableDB [func] (db_test.go)
├── TestNewPool_UnreachableDB [func] (db_test.go)
└── testLogger [func] (db_test.go)

