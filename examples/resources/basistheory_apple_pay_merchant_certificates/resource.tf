resource "basistheory_apple_pay_merchant_registration" "example" {
  merchant_identifier = "merchant.com.example.app"
}

resource "basistheory_apple_pay_merchant_certificates" "example" {
  merchant_registration_id               = basistheory_apple_pay_merchant_registration.example.id
  merchant_certificate_data              = filebase64("certs/merchant.p12")
  merchant_certificate_password          = basistheory_google_pay_merchant_registration.merchant.password
  payment_processor_certificate_data     = filebase64("certs/payment-processor.p12")
  payment_processor_certificate_password = basistheory_google_pay_merchant_registration.payment_processor.password
  domain                                 = "checkout.example.com"
}
