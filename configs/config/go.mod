module github.com/mejzh77/astragen/configs/config

go 1.24.4

replace github.com/mejzh77/astragen/pkg/models => ../../pkg/models

require (
	github.com/mejzh77/astragen/pkg/models v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.20.0 // indirect
	gorm.io/gorm v1.30.0 // indirect
)
