// TODO(sam): pull this in from docker environent
db_url = "postgres://democart:democartpass@db/democart?sslmode=disable"
//db_url = "sqlite3:democart.sqlite3.db?sslmode=disable"

server_url = "http://localhost:8080"
client_hosts = ["http://localhost:3000"]

graceful_shutdown_timeout_sec = 5
write_timeout_sec             = 15
read_timeout_sec              = 15
idle_timeout_sec              = 15

idp_password_salt = "00000"
idp_client_id     = "idp_client_id"
idp_client_secret = "idp_client_secret"

loglevel = "debug"
developer_mode = true
