data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./",
    "--dialect", "postgres",
  ]
}
env "gorm" {
  src = data.external_schema.gorm.url
  url = "${TARGET_DB_DSN}"
  dev = "${DEV_DB_DSN}"
  migration {
    dir = "file://../.ci/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}