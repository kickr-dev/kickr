terraform { backend "http" {} }
variable "my_secret" { sensitive = true }
variable "github_var" {}
variable "my_var" {}
