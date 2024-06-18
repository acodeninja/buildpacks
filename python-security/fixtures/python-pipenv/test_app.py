import xmlsec

def test_sign():
    data = b'\xa8f4dP\x82\x02\xd3\xf5.\x02\xc1\x03\xef\xc4\x86\xabC\xec\xb7>\x8e\x1f\xa3\xa3\xc5\xb9qc\xc2\x81\xb1-\xa4B\xdf\x03>\xba\xd1'

    ctx = xmlsec.SignatureContext()
    key = xmlsec.Key.from_file('keys/id_rsa.pem', xmlsec.constants.KeyDataFormatPem)
    ctx.key = key
    sign = ctx.sign_binary(data, xmlsec.constants.TransformRsaSha1)

    print(sign)
