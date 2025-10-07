data "external_schema" "gorm" {
  program = [
  "go",
  "run",
  "-mod=mod",
  "ariga.io/atlas-provider-gorm",
  "load",
  "--path", "./path/to/models",
  "--dialect", "sqlite", // | mysql | sqlite | sqlserver
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "sqlite://file?mode=memory"
}
