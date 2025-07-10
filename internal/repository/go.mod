module github.com/mejzh77/astragen/internal/repository

go 1.24.4

require (
	github.com/mejzh77/astragen/pkg/models v0.0.0-00010101000000-000000000000
	gorm.io/gorm v1.30.0
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/mejzh77/astragen/pkg/models => ../../pkg/models
