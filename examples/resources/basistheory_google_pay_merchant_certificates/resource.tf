resource "basistheory_google_pay_merchant_registration" "example" {
  merchant_identifier = "BCR2DN4TQDNXXX5YZ"
}

resource "basistheory_google_pay_merchant_certificates" "example" {
  merchant_registration_id      = basistheory_google_pay_merchant_registration.example.id
  merchant_certificate_data     = filebase64("certs/merchant.p12")
  merchant_certificate_password = var.merchant_certificate_password
}
