# Signed contaieners using Bazel and Github workflows 

Simple repo that builds a go application in github and then uses `cosign` to sign the image.

### References

- [Deterministic container hashes and container signing using Cosign, Bazel and Google Cloud Build](https://github.com/salrashid123/cosign_bazel_cloud_build)
- [Generate and verify cosign signatures using openssl](https://gist.github.com/salrashid123/4138d8b6ed5a89f5569e44eecbbb8fda)

Note, the cosign sample is attributed to [WFA Measurement System](https://github.com/world-federation-of-advertisers/cross-media-measurement)

### Build and Test locally

To build and test the image locally
```bash
bazel run :gazelle -- update-repos -from_file=go.mod -prune=true -to_macro=repositories.bzl%go_repositories

bazel run app:server -- -httpport :8080
bazelisk run app:server -- -httpport :8080

bazel test app:go_default_test
bazelisk test app:go_default_test

go test -v ./...
```

### Push the image to DockerHub

To push the image locally
```bash
bazelisk run app:push-image

export IMAGE="docker.io/salrashid123/server_image:server@sha256:d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5"

crane  manifest $IMAGE
```

### Sign the image on dockerhub

Sign locally

```bash
bazel run app:sign_all_images --define container_registry=docker.io --define image_repo_prefix=salrashid123 --define image_tag=serverx
```

### Verify the image

```bash
export COSIGN_EXPERIMENTAL=1  
export IMAGE="docker.io/salrashid123/server_image:server@sha256:d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5"

cosign verify $IMAGE     --certificate-oidc-issuer https://token.actions.githubusercontent.com       --certificate-identity-regexp="https://github.com.*"

cosign verify $IMAGE     --certificate-oidc-issuer https://github.com/login/oauth   --certificate-identity salrashid123@gmail.com 

cosign verify $IMAGE   --certificate-identity salrashid123@gmail.com  --certificate-oidc-issuer https://accounts.google.com
```

---

### Trace

```bash
$ cosign verify $IMAGE     --certificate-oidc-issuer https://github.com/login/oauth   --certificate-identity salrashid123@gmail.com  | jq '.'

Verification for index.docker.io/salrashid123/server_image@sha256:d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5 --

[
  {
    "critical": {
      "identity": {
        "docker-reference": "index.docker.io/salrashid123/server_image"
      },
      "image": {
        "docker-manifest-digest": "sha256:d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5"
      },
      "type": "cosign container image signature"
    },
    "optional": {
      "1.3.6.1.4.1.57264.1.1": "https://github.com/login/oauth",
      "Bundle": {
        "SignedEntryTimestamp": "MEUCIAQNhYJRzvfgL6XDX5Nk9jToIp1NWWJRebRX3BE6RsXjAiEA6QbQtWAzVV7ykxyNxlSy8eQ5BYw5EVVPiUGTePhR2yM=",
        "Payload": {
          "body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI2YmUxNGFmYmFjYjZhNzRmNjUxZmUwMWY2YjQ2YzYzMmRkYjUyOTU0ZDg2YmZjYjQ5YWIyMTE4MWMwYzE5NWI3In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FWUNJUURRZ3VEZUkzbm55WnZwWnNtSE9UdmR5eFROd1FpZUNMUFgxZlF2Y0p3a0VRSWhBT01TZU9mYkk1UkRPZ2lWbmVFQkgvOHdaUU9UcXNOdEhsT1JwT3RxUkRWUSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVTXhWRU5EUVd4MVowRjNTVUpCWjBsVlVVaFlUMGxSZW5od2FrRkdkVmRWVUdveU1FaEhXR1ZxY2xWM2QwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUVRGTlZFa3hUa1JSTVZkb1kwNU5hbGwzVFdwQk1VMVVUWGRPUkZFeFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVZLTjNGWVJtTnBlbVZZU1hJMWMydHlhaTh4UlZaRFUzaHNTMjgwVkZCc0szbEhNbWtLYzFOd1QybzVaMFZFYVZvNE1uSkNRM050YmtacVdTdHdRMEZCT1dSUFUwMVdLMUozYTJSRWJsZHhOM2xXTTNaRk9IRlBRMEZZYjNkblowWXlUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlV4VXpoWENsVllTVlpOVERrNGVYVkdOVlU0VTBaSldrODJUVUZCZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDBwQldVUldVakJTUVZGSUwwSkNiM2RIU1VWWFl6SkdjMk50Um5waFIyeHJUVlJKZWxGSFpIUlpWMnh6VEcxT2RtSlVRWE5DWjI5eVFtZEZSUXBCV1U4dlRVRkZRa0pDTlc5a1NGSjNZM3B2ZGt3eVpIQmtSMmd4V1drMWFtSXlNSFppUnpsdVlWYzBkbUl5UmpGa1IyZDNUR2RaUzB0M1dVSkNRVWRFQ25aNlFVSkRRVkZuUkVJMWIyUklVbmRqZW05MlRESmtjR1JIYURGWmFUVnFZakl3ZG1KSE9XNWhWelIyWWpKR01XUkhaM2RuV1c5SFEybHpSMEZSVVVJS01XNXJRMEpCU1VWbVFWSTJRVWhuUVdSblJHUlFWRUp4ZUhOalVrMXRUVnBJYUhsYVducGpRMjlyY0dWMVRqUTRjbVlyU0dsdVMwRk1lVzUxYW1kQlFRcEJXbmQwTTI5d2NrRkJRVVZCZDBKSVRVVlZRMGxIZFdKRFVXUjNiSHBwZFZGdmNEQmFXRU5JVVZaSE1uSkhWMjh6Vmt0UmFTdExOR2xOTUVoSlprRlNDa0ZwUlVGc1VVaGpOblkwWVdreVVVMXRWVlpNYUhaeVJXeHBhR1E1UTA0eFQxcHZSRUZXWTBGMmRETXlkbTF2ZDBObldVbExiMXBKZW1vd1JVRjNUVVFLWVVGQmQxcFJTWGhCVGtadWJ6TlhRMGhJTkhZNU0xUkJLMWhHTkUxQ1ZEWnNSV1I1YVVWamRFOUROSFp5ZEVGa1pWTlFZWFl5Y2poUE1sSXZVSEoxY3dvd1dWWlRSRVpsSzB0blNYZElXbmMxU2tWellscHdVbGdyWldGdlVFMHZRMEY1ZW1ZMmNrdDFSa293YldFMVJFTkdZM1pIYlcxc1oyaDNWVEpES3paWkNpdG5aVGREZFdaeGNWcDBhZ290TFMwdExVVk9SQ0JEUlZKVVNVWkpRMEZVUlMwdExTMHRDZz09In19fX0=",
          "integratedTime": 1770296109,
          "logIndex": 919437390,
          "logID": "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
        }
      },
      "Issuer": "https://github.com/login/oauth",
      "Subject": "salrashid123@gmail.com"
    }
  }
]
```

```json
{
  "apiVersion": "0.0.1",
  "kind": "hashedrekord",
  "spec": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "6be14afbacb6a74f651fe01f6b46c632ddb52954d86bfcb49ab21181c0c195b7"
      }
    },
    "signature": {
      "content": "MEYCIQDQguDeI3nnyZvpZsmHOTvdyxTNwQieCLPX1fQvcJwkEQIhAOMSeOfbI5RDOgiVneEBH/8wZQOTqsNtHlORpOtqRDVQ",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMxVENDQWx1Z0F3SUJBZ0lVUUhYT0lRenhwakFGdVdVUGoyMEhHWGVqclV3d0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qQTFNVEkxTkRRMVdoY05Nall3TWpBMU1UTXdORFExV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVKN3FYRmNpemVYSXI1c2tyai8xRVZDU3hsS280VFBsK3lHMmkKc1NwT2o5Z0VEaVo4MnJCQ3NtbkZqWStwQ0FBOWRPU01WK1J3a2REbldxN3lWM3ZFOHFPQ0FYb3dnZ0YyTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVUxUzhXClVYSVZNTDk4eXVGNVU4U0ZJWk82TUFBd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d0pBWURWUjBSQVFIL0JCb3dHSUVXYzJGc2NtRnphR2xrTVRJelFHZHRZV2xzTG1OdmJUQXNCZ29yQmdFRQpBWU8vTUFFQkJCNW9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZiRzluYVc0dmIyRjFkR2d3TGdZS0t3WUJCQUdECnZ6QUJDQVFnREI1b2RIUndjem92TDJkcGRHaDFZaTVqYjIwdmJHOW5hVzR2YjJGMWRHZ3dnWW9HQ2lzR0FRUUIKMW5rQ0JBSUVmQVI2QUhnQWRnRGRQVEJxeHNjUk1tTVpIaHlaWnpjQ29rcGV1TjQ4cmYrSGluS0FMeW51amdBQQpBWnd0M29wckFBQUVBd0JITUVVQ0lHdWJDUWR3bHppdVFvcDBaWENIUVZHMnJHV28zVktRaStLNGlNMEhJZkFSCkFpRUFsUUhjNnY0YWkyUU1tVVZMaHZyRWxpaGQ5Q04xT1pvREFWY0F2dDMydm1vd0NnWUlLb1pJemowRUF3TUQKYUFBd1pRSXhBTkZubzNXQ0hINHY5M1RBK1hGNE1CVDZsRWR5aUVjdE9DNHZydEFkZVNQYXYycjhPMlIvUHJ1cwowWVZTREZlK0tnSXdIWnc1SkVzYlpwUlgrZWFvUE0vQ0F5emY2ckt1RkowbWE1RENGY3ZHbW1sZ2h3VTJDKzZZCitnZTdDdWZxcVp0agotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
      }
    }
  }
}
```

```bash
cat c.crt

-----BEGIN CERTIFICATE-----
MIIC1TCCAlugAwIBAgIUQHXOIQzxpjAFuWUPj20HGXejrUwwCgYIKoZIzj0EAwMw
NzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl
cm1lZGlhdGUwHhcNMjYwMjA1MTI1NDQ1WhcNMjYwMjA1MTMwNDQ1WjAAMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAEJ7qXFcizeXIr5skrj/1EVCSxlKo4TPl+yG2i
sSpOj9gEDiZ82rBCsmnFjY+pCAA9dOSMV+RwkdDnWq7yV3vE8qOCAXowggF2MA4G
A1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU1S8W
UXIVML98yuF5U8SFIZO6MAAwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y
ZD8wJAYDVR0RAQH/BBowGIEWc2FscmFzaGlkMTIzQGdtYWlsLmNvbTAsBgorBgEE
AYO/MAEBBB5odHRwczovL2dpdGh1Yi5jb20vbG9naW4vb2F1dGgwLgYKKwYBBAGD
vzABCAQgDB5odHRwczovL2dpdGh1Yi5jb20vbG9naW4vb2F1dGgwgYoGCisGAQQB
1nkCBAIEfAR6AHgAdgDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+HinKALynujgAA
AZwt3oprAAAEAwBHMEUCIGubCQdwlziuQop0ZXCHQVG2rGWo3VKQi+K4iM0HIfAR
AiEAlQHc6v4ai2QMmUVLhvrElihd9CN1OZoDAVcAvt32vmowCgYIKoZIzj0EAwMD
aAAwZQIxANFno3WCHH4v93TA+XF4MBT6lEdyiEctOC4vrtAdeSPav2r8O2R/Prus
0YVSDFe+KgIwHZw5JEsbZpRX+eaoPM/CAyzf6rKuFJ0ma5DCFcvGmmlghwU2C+6Y
+ge7CufqqZtj
-----END CERTIFICATE-----


$  openssl x509 -in c.crt -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            40:75:ce:21:0c:f1:a6:30:05:b9:65:0f:8f:6d:07:19:77:a3:ad:4c
        Signature Algorithm: ecdsa-with-SHA384
        Issuer: O=sigstore.dev, CN=sigstore-intermediate
        Validity
            Not Before: Feb  5 12:54:45 2026 GMT
            Not After : Feb  5 13:04:45 2026 GMT
        Subject: 
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:27:ba:97:15:c8:b3:79:72:2b:e6:c9:2b:8f:fd:
                    44:54:24:b1:94:aa:38:4c:f9:7e:c8:6d:a2:b1:2a:
                    4e:8f:d8:04:0e:26:7c:da:b0:42:b2:69:c5:8d:8f:
                    a9:08:00:3d:74:e4:8c:57:e4:70:91:d0:e7:5a:ae:
                    f2:57:7b:c4:f2
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage: 
                Code Signing
            X509v3 Subject Key Identifier: 
                D5:2F:16:51:72:15:30:BF:7C:CA:E1:79:53:C4:85:21:93:BA:30:00
            X509v3 Authority Key Identifier: 
                DF:D3:E9:CF:56:24:11:96:F9:A8:D8:E9:28:55:A2:C6:2E:18:64:3F
            X509v3 Subject Alternative Name: critical
                email:salrashid123@gmail.com
            1.3.6.1.4.1.57264.1.1: 
                https://github.com/login/oauth
            1.3.6.1.4.1.57264.1.8: 
                ..https://github.com/login/oauth
            CT Precertificate SCTs: 
                Signed Certificate Timestamp:
                    Version   : v1 (0x0)
                    Log ID    : DD:3D:30:6A:C6:C7:11:32:63:19:1E:1C:99:67:37:02:
                                A2:4A:5E:B8:DE:3C:AD:FF:87:8A:72:80:2F:29:EE:8E
                    Timestamp : Feb  5 12:54:45.099 2026 GMT
                    Extensions: none
                    Signature : ecdsa-with-SHA256
                                30:45:02:20:6B:9B:09:07:70:97:38:AE:42:8A:74:65:
                                70:87:41:51:B6:AC:65:A8:DD:52:90:8B:E2:B8:88:CD:
                                07:21:F0:11:02:21:00:95:01:DC:EA:FE:1A:8B:64:0C:
                                99:45:4B:86:FA:C4:96:28:5D:F4:23:75:39:9A:03:01:
                                57:00:BE:DD:F6:BE:6A
    Signature Algorithm: ecdsa-with-SHA384
    Signature Value:
        30:65:02:31:00:d1:67:a3:75:82:1c:7e:2f:f7:74:c0:f9:71:
        78:30:14:fa:94:47:72:88:47:2d:38:2e:2f:ae:d0:1d:79:23:
        da:bf:6a:fc:3b:64:7f:3e:bb:ac:d1:85:52:0c:57:be:2a:02:
        30:1d:9c:39:24:4b:1b:66:94:57:f9:e6:a8:3c:cf:c2:03:2c:
        df:ea:b2:ae:14:9d:26:6b:90:c2:15:cb:c6:9a:69:60:87:05:
        36:0b:ee:98:fa:07:bb:0a:e7:ea:a9:9b:63

```

```bash
$ rekor-cli search --rekor_server https://rekor.sigstore.dev    --sha  6be14afbacb6a74f651fe01f6b46c632ddb52954d86bfcb49ab21181c0c195b7
Found matching entries (listed by UUID):
108e9186e8c5677af0a3d34d316d6d83e5dbc1f4f6166e019c5499e57ca2be40f4f09b7c152dafb7

$ rekor-cli get --rekor_server https://rekor.sigstore.dev    --uuid 108e9186e8c5677af0a3d34d316d6d83e5dbc1f4f6166e019c5499e57ca2be40f4f09b7c152dafb7 
LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 919437390
IntegratedTime: 2026-02-05T12:55:09Z
UUID: 108e9186e8c5677af0a3d34d316d6d83e5dbc1f4f6166e019c5499e57ca2be40f4f09b7c152dafb7
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "6be14afbacb6a74f651fe01f6b46c632ddb52954d86bfcb49ab21181c0c195b7"
      }
    },
    "signature": {
      "content": "MEYCIQDQguDeI3nnyZvpZsmHOTvdyxTNwQieCLPX1fQvcJwkEQIhAOMSeOfbI5RDOgiVneEBH/8wZQOTqsNtHlORpOtqRDVQ",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMxVENDQWx1Z0F3SUJBZ0lVUUhYT0lRenhwakFGdVdVUGoyMEhHWGVqclV3d0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qQTFNVEkxTkRRMVdoY05Nall3TWpBMU1UTXdORFExV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVKN3FYRmNpemVYSXI1c2tyai8xRVZDU3hsS280VFBsK3lHMmkKc1NwT2o5Z0VEaVo4MnJCQ3NtbkZqWStwQ0FBOWRPU01WK1J3a2REbldxN3lWM3ZFOHFPQ0FYb3dnZ0YyTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVUxUzhXClVYSVZNTDk4eXVGNVU4U0ZJWk82TUFBd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d0pBWURWUjBSQVFIL0JCb3dHSUVXYzJGc2NtRnphR2xrTVRJelFHZHRZV2xzTG1OdmJUQXNCZ29yQmdFRQpBWU8vTUFFQkJCNW9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZiRzluYVc0dmIyRjFkR2d3TGdZS0t3WUJCQUdECnZ6QUJDQVFnREI1b2RIUndjem92TDJkcGRHaDFZaTVqYjIwdmJHOW5hVzR2YjJGMWRHZ3dnWW9HQ2lzR0FRUUIKMW5rQ0JBSUVmQVI2QUhnQWRnRGRQVEJxeHNjUk1tTVpIaHlaWnpjQ29rcGV1TjQ4cmYrSGluS0FMeW51amdBQQpBWnd0M29wckFBQUVBd0JITUVVQ0lHdWJDUWR3bHppdVFvcDBaWENIUVZHMnJHV28zVktRaStLNGlNMEhJZkFSCkFpRUFsUUhjNnY0YWkyUU1tVVZMaHZyRWxpaGQ5Q04xT1pvREFWY0F2dDMydm1vd0NnWUlLb1pJemowRUF3TUQKYUFBd1pRSXhBTkZubzNXQ0hINHY5M1RBK1hGNE1CVDZsRWR5aUVjdE9DNHZydEFkZVNQYXYycjhPMlIvUHJ1cwowWVZTREZlK0tnSXdIWnc1SkVzYlpwUlgrZWFvUE0vQ0F5emY2ckt1RkowbWE1RENGY3ZHbW1sZ2h3VTJDKzZZCitnZTdDdWZxcVp0agotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
      }
    }
  }
}
```


```bash
$ crane  manifest salrashid123/server_image:sha256-d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5.sig | jq '.'


{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 233,
    "digest": "sha256:e7e4e8dc998d0300223919a0baba07e48ecbdb2a6bd2798b01fac63f6bae3e29"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:6be14afbacb6a74f651fe01f6b46c632ddb52954d86bfcb49ab21181c0c195b7",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEYCIQDQguDeI3nnyZvpZsmHOTvdyxTNwQieCLPX1fQvcJwkEQIhAOMSeOfbI5RDOgiVneEBH/8wZQOTqsNtHlORpOtqRDVQ",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEUCIAQNhYJRzvfgL6XDX5Nk9jToIp1NWWJRebRX3BE6RsXjAiEA6QbQtWAzVV7ykxyNxlSy8eQ5BYw5EVVPiUGTePhR2yM=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI2YmUxNGFmYmFjYjZhNzRmNjUxZmUwMWY2YjQ2YzYzMmRkYjUyOTU0ZDg2YmZjYjQ5YWIyMTE4MWMwYzE5NWI3In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FWUNJUURRZ3VEZUkzbm55WnZwWnNtSE9UdmR5eFROd1FpZUNMUFgxZlF2Y0p3a0VRSWhBT01TZU9mYkk1UkRPZ2lWbmVFQkgvOHdaUU9UcXNOdEhsT1JwT3RxUkRWUSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVTXhWRU5EUVd4MVowRjNTVUpCWjBsVlVVaFlUMGxSZW5od2FrRkdkVmRWVUdveU1FaEhXR1ZxY2xWM2QwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUVRGTlZFa3hUa1JSTVZkb1kwNU5hbGwzVFdwQk1VMVVUWGRPUkZFeFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVZLTjNGWVJtTnBlbVZZU1hJMWMydHlhaTh4UlZaRFUzaHNTMjgwVkZCc0szbEhNbWtLYzFOd1QybzVaMFZFYVZvNE1uSkNRM050YmtacVdTdHdRMEZCT1dSUFUwMVdLMUozYTJSRWJsZHhOM2xXTTNaRk9IRlBRMEZZYjNkblowWXlUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlV4VXpoWENsVllTVlpOVERrNGVYVkdOVlU0VTBaSldrODJUVUZCZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDBwQldVUldVakJTUVZGSUwwSkNiM2RIU1VWWFl6SkdjMk50Um5waFIyeHJUVlJKZWxGSFpIUlpWMnh6VEcxT2RtSlVRWE5DWjI5eVFtZEZSUXBCV1U4dlRVRkZRa0pDTlc5a1NGSjNZM3B2ZGt3eVpIQmtSMmd4V1drMWFtSXlNSFppUnpsdVlWYzBkbUl5UmpGa1IyZDNUR2RaUzB0M1dVSkNRVWRFQ25aNlFVSkRRVkZuUkVJMWIyUklVbmRqZW05MlRESmtjR1JIYURGWmFUVnFZakl3ZG1KSE9XNWhWelIyWWpKR01XUkhaM2RuV1c5SFEybHpSMEZSVVVJS01XNXJRMEpCU1VWbVFWSTJRVWhuUVdSblJHUlFWRUp4ZUhOalVrMXRUVnBJYUhsYVducGpRMjlyY0dWMVRqUTRjbVlyU0dsdVMwRk1lVzUxYW1kQlFRcEJXbmQwTTI5d2NrRkJRVVZCZDBKSVRVVlZRMGxIZFdKRFVXUjNiSHBwZFZGdmNEQmFXRU5JVVZaSE1uSkhWMjh6Vmt0UmFTdExOR2xOTUVoSlprRlNDa0ZwUlVGc1VVaGpOblkwWVdreVVVMXRWVlpNYUhaeVJXeHBhR1E1UTA0eFQxcHZSRUZXWTBGMmRETXlkbTF2ZDBObldVbExiMXBKZW1vd1JVRjNUVVFLWVVGQmQxcFJTWGhCVGtadWJ6TlhRMGhJTkhZNU0xUkJLMWhHTkUxQ1ZEWnNSV1I1YVVWamRFOUROSFp5ZEVGa1pWTlFZWFl5Y2poUE1sSXZVSEoxY3dvd1dWWlRSRVpsSzB0blNYZElXbmMxU2tWellscHdVbGdyWldGdlVFMHZRMEY1ZW1ZMmNrdDFSa293YldFMVJFTkdZM1pIYlcxc1oyaDNWVEpES3paWkNpdG5aVGREZFdaeGNWcDBhZ290TFMwdExVVk9SQ0JEUlZKVVNVWkpRMEZVUlMwdExTMHRDZz09In19fX0=\",\"integratedTime\":1770296109,\"logIndex\":919437390,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
        "dev.sigstore.cosign/certificate": "-----BEGIN CERTIFICATE-----\nMIIC1TCCAlugAwIBAgIUQHXOIQzxpjAFuWUPj20HGXejrUwwCgYIKoZIzj0EAwMw\nNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl\ncm1lZGlhdGUwHhcNMjYwMjA1MTI1NDQ1WhcNMjYwMjA1MTMwNDQ1WjAAMFkwEwYH\nKoZIzj0CAQYIKoZIzj0DAQcDQgAEJ7qXFcizeXIr5skrj/1EVCSxlKo4TPl+yG2i\nsSpOj9gEDiZ82rBCsmnFjY+pCAA9dOSMV+RwkdDnWq7yV3vE8qOCAXowggF2MA4G\nA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU1S8W\nUXIVML98yuF5U8SFIZO6MAAwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y\nZD8wJAYDVR0RAQH/BBowGIEWc2FscmFzaGlkMTIzQGdtYWlsLmNvbTAsBgorBgEE\nAYO/MAEBBB5odHRwczovL2dpdGh1Yi5jb20vbG9naW4vb2F1dGgwLgYKKwYBBAGD\nvzABCAQgDB5odHRwczovL2dpdGh1Yi5jb20vbG9naW4vb2F1dGgwgYoGCisGAQQB\n1nkCBAIEfAR6AHgAdgDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+HinKALynujgAA\nAZwt3oprAAAEAwBHMEUCIGubCQdwlziuQop0ZXCHQVG2rGWo3VKQi+K4iM0HIfAR\nAiEAlQHc6v4ai2QMmUVLhvrElihd9CN1OZoDAVcAvt32vmowCgYIKoZIzj0EAwMD\naAAwZQIxANFno3WCHH4v93TA+XF4MBT6lEdyiEctOC4vrtAdeSPav2r8O2R/Prus\n0YVSDFe+KgIwHZw5JEsbZpRX+eaoPM/CAyzf6rKuFJ0ma5DCFcvGmmlghwU2C+6Y\n+ge7CufqqZtj\n-----END CERTIFICATE-----\n",
        "dev.sigstore.cosign/chain": "-----BEGIN CERTIFICATE-----\nMIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRkHrn/UnfaQwCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMjA0MTMyMDA2MTVaFw0zMTEwMDUxMzU2NThaMDcxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjEeMBwGA1UEAxMVc2lnc3RvcmUtaW50ZXJtZWRpYXRlMHYwEAYHKoZIzj0C\nAQYFK4EEACIDYgAE8RVS/ysH+NOvuDZyPIZtilgUF9NlarYpAd9HP1vBBH1U5CV7\n7LSS7s0ZiH4nE7Hv7ptS6LvvR/STk798LVgMzLlJ4HeIfF3tHSaexLcYpSASr1kS\n0N/RgBJz/9jWCiXno3sweTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYB\nBQUHAwMwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU39Ppz1YkEZb5qNjp\nKFWixi4YZD8wHwYDVR0jBBgwFoAUWMAeX5FFpWapesyQoZMi0CrFxfowCgYIKoZI\nzj0EAwMDZwAwZAIwPCsQK4DYiZYDPIaDi5HFKnfxXx6ASSVmERfsynYBiX2X6SJR\nnZU84/9DZdnFvvxmAjBOt6QpBlc4J/0DxvkTCqpclvziL6BCCPnjdlIB3Pu3BxsP\nmygUY7Ii2zbdCdliiow=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIB9zCCAXygAwIBAgIUALZNAPFdxHPwjeDloDwyYChAO/4wCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMTEwMDcxMzU2NTlaFw0zMTEwMDUxMzU2NThaMCoxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjERMA8GA1UEAxMIc2lnc3RvcmUwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT7\nXeFT4rb3PQGwS4IajtLk3/OlnpgangaBclYpsYBr5i+4ynB07ceb3LP0OIOZdxex\nX69c5iVuyJRQ+Hz05yi+UF3uBWAlHpiS5sh0+H2GHE7SXrk1EC5m1Tr19L9gg92j\nYzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRY\nwB5fkUWlZql6zJChkyLQKsXF+jAfBgNVHSMEGDAWgBRYwB5fkUWlZql6zJChkyLQ\nKsXF+jAKBggqhkjOPQQDAwNpADBmAjEAj1nHeXZp+13NWBNa+EDsDP8G1WWg1tCM\nWP/WHPqpaVo0jhsweNFZgSs0eE7wYI4qAjEA2WB9ot98sIkoF3vZYdd3/VtWB5b9\nTNMea7Ix/stJ5TfcLLeABLE4BNJOsQ4vnBHJ\n-----END CERTIFICATE-----"
      }
    }
  ]
}
```