resource "basistheory_apple_pay_merchant_registration" "example" {
  merchant_identifier = "merchant.com.example.app"
}

resource "basistheory_apple_pay_merchant_certificates" "example" {
  merchant_registration_id               = basistheory_apple_pay_merchant_registration.example.id
  merchant_certificate_data              = filebase64("certs/merchant.p12")
  merchant_certificate_password          = var.merchant_certificate_password
  payment_processor_certificate_data     = filebase64("certs/payment-processor.p12")
  payment_processor_certificate_password = var.payment_processor_certificate_password
  domain                                 = "checkout.example.com"
}
