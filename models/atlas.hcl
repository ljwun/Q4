variable "target_db" {
  type    = string
  default = getenv("TARGET_DB_DSN")
}

variable "dev_db" {
  type    = string
  default = getenv("DEV_DB_DSN")
}

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./models/",
    "--dialect", "postgres",
  ]
}
env "gorm" {
  src = data.external_schema.gorm.url
  url = var.target_db
  dev = var.dev_db
  migration {
    dir = "file://.ci/migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}