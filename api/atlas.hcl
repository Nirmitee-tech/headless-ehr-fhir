env "local" {
  src = "file://migrations"
  url = "postgres://ehr:changeme@localhost:5432/ehr?sslmode=disable&search_path=public"

  migration {
    dir = "file://migrations"
  }
}

env "docker" {
  src = "file://migrations"
  url = env("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}
