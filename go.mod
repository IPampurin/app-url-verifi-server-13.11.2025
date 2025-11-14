module github.com/IPampurin/app-url-verifi-server-13.11.2025

go 1.24.1

replace verifi-server => ./

require (
	github.com/joho/godotenv v1.5.1
	github.com/jung-kurt/gofpdf v1.16.2
	verifi-server v0.0.0-00010101000000-000000000000
)
