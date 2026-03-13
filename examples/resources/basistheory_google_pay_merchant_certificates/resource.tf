resource "basistheory_google_pay_merchant_registration" "example" {
  merchant_identifier = "your_merchant"
}

resource "basistheory_google_pay_merchant_certificates" "example" {
  merchant_registration_id      = basistheory_google_pay_merchant_registration.example.id
  merchant_certificate_data     = filebase64("certs/merchant.p12")
  merchant_certificate_password = basistheory_google_pay_merchant_registration.merchant.password
}
