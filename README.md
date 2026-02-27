# Signed containers using Bazel and Github workflows 

Simple repo which:

1. builds a go application using `bazel`
2. creates an oci image using `bazel` 
3. pushes the image to dockerhub
4. uses `cosign sign` to sign the image and add an entry to sigstore transparency log
5. uses `cosign sign-blob` for each binary in the releases page.

All this is done within a github workflow so its auditable end-to-end

### References

- [Deterministic container hashes and container signing using Cosign, Bazel and Google Cloud Build](https://github.com/salrashid123/cosign_bazel_cloud_build)
- [Generate and verify cosign signatures using openssl](https://gist.github.com/salrashid123/4138d8b6ed5a89f5569e44eecbbb8fda)

Note, the cosign sample is attributed to [WFA Measurement System](https://github.com/world-federation-of-advertisers/cross-media-measurement)

### Build and Test locally

To build and test the image locally

```bash
bazel run :gazelle -- update-repos -from_file=go.mod -prune=true -to_macro=repositories.bzl%go_repositories

#bazel run app:server -- -httpport :8080
bazelisk run app:server -- -httpport :8080

#bazel test app:go_default_test
bazelisk test app:go_default_test

# or using go directly
go test -v ./...
```

### Push the image to DockerHub

To push the image locally
```bash
bazelisk run app:push-image

export IMAGE="docker.io/salrashid123/server_image:server@sha256:8c0d347ab4b606c93ba087efc5c0e6e58154b052b0eb1d658a313cdaeff93d71"

crane  manifest $IMAGE
```

Default image is posted to

* [https://hub.docker.com/r/salrashid123/server_image/tags](https://hub.docker.com/r/salrashid123/server_image/tags)

### Sign the image on dockerhub

If you want ot sign locally using bazel


```bash
bazelisk run app:sign_all_images --define container_registry=docker.io --define image_repo_prefix=salrashid123 --define image_tag=server -- -y
```

see [cosign sign](https://docs.sigstore.dev/cosign/signing/overview/)


### Verify the image

You can use `cosign` to  verify locally too:

```bash
export COSIGN_EXPERIMENTAL=1  
export IMAGE="docker.io/salrashid123/server_image:server@sha256:8c0d347ab4b606c93ba087efc5c0e6e58154b052b0eb1d658a313cdaeff93d71"

# for github workflow
cosign verify $IMAGE     --certificate-oidc-issuer https://token.actions.githubusercontent.com       --certificate-identity-regexp="https://github.com.*"

# for local
# cosign verify $IMAGE     --certificate-oidc-issuer https://github.com/login/oauth   --certificate-identity salrashid123@gmail.com 

# for google oidc
# cosign verify $IMAGE   --certificate-identity salrashid123@gmail.com  --certificate-oidc-issuer https://accounts.google.com
```

see [Verify image with user-provided trusted chain](https://docs.sigstore.dev/cosign/verifying/verify/)


### Github Workflow

If you want to test the full end to end workflow on github, you need to setup the github token and gpg signatures as shown in the `.github/workflows/` folder

```bash
git add -A
git commit -m "add code" -S -s
git push

export TAG=v0.0.28
git tag -a $TAG -m "Release $TAG" -s
git push origin $TAG
```

![images/cosign_workflo.png](images/cosign_workflow.png)

---

### Github Attestation

The github workflow also uses `goreleaser` to generate binarires (you can use bazel but i wanted to generate the [attestations](https://github.com/salrashid123/go-bazel-github-workflow/attestations/18349640) easily)

![images/attestation.png](images/attestation.png)

### Trace

#### Verify container image

Once the code is pushed, you can recall the entire signature from the log.

```bash
cosign verify $IMAGE     --certificate-oidc-issuer https://token.actions.githubusercontent.com   --certificate-identity-regexp="https://github.com.*"  | jq '.'

Verification for index.docker.io/salrashid123/server_image@sha256:8c0d347ab4b606c93ba087efc5c0e6e58154b052b0eb1d658a313cdaeff93d71 --
The following checks were performed on each of these signatures:
  - The cosign claims were validated
  - Existence of the claims in the transparency log was verified offline
  - The code-signing certificate was verified using trusted certificate authority certificates
[
  {
    "critical": {
      "identity": {
        "docker-reference": "index.docker.io/salrashid123/server_image"
      },
      "image": {
        "docker-manifest-digest": "sha256:8c0d347ab4b606c93ba087efc5c0e6e58154b052b0eb1d658a313cdaeff93d71"
      },
      "type": "cosign container image signature"
    },
    "optional": {
      "1.3.6.1.4.1.57264.1.1": "https://token.actions.githubusercontent.com",
      "1.3.6.1.4.1.57264.1.2": "push",
      "1.3.6.1.4.1.57264.1.3": "15dcefe0d3785f38673fff224da0bddf69f40e04",
      "1.3.6.1.4.1.57264.1.4": "Release",
      "1.3.6.1.4.1.57264.1.5": "salrashid123/go-bazel-github-workflow",
      "1.3.6.1.4.1.57264.1.6": "refs/tags/v0.0.28",
      "Bundle": {
        "SignedEntryTimestamp": "MEUCIEMK6jfMz5mP8QP1hehjMOQzZnW7JaN56h/P5LprNF3LAiEAz/upDohXVl+EediGiPEL7mvVacbIGKHlrhVSw0Q4CXw=",
        "Payload": {
          "body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI1NjE0ZGY5ZDBlMWZkODBlOGE0YWIyZjBhZjQ3ZmNkYjRmOTRiMjhiNGRlYWQxZTRiOWM3OGJkYTMxMjA5YThlIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FWUNJUURzWkpBTUtPMWorbUt3b1p6UUpDV0Y1RFdQOTF0Mkw0cVcwZXU1Vy9lVlRnSWhBUHpadjBTUWpxNzQ1TXcyUmhtWUNSRFlSS0toenNsRFZDQys5aldMaVFTayIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1ha05EUW5KWFowRjNTVUpCWjBsVlVrbE9kRGhXVVVSbGRVOXZjRmw2VUVacFdWTjJPRXhFTm5GbmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxU1ROTlZFVXdUbXBSTkZkb1kwNU5hbGwzVFdwSk0wMVVSVEZPYWxFMFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVY1WTI0MlNuTm5TU3N5UlcxUWVsTkdORzFuZVc5cGFWSk5TRmxKU21ObVlVaEpOamNLT0ROMlNYcENOVTVVYkVadlFqZDZUMnQ2Y3pSeFQwOHdZa3RqV2xRNWVsa3lTMkZsYjNNNE4yd3ZhV0ZKUlRabkwyRlBRMEprVVhkbloxaFJUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlV2YTBWV0NtZGtNVmxvU2t0UlNIZHJaSGhNTHpoRFMxWkViek4zZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTkUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRFozaE9WMUpxV2xkYWJFMUhVWHBPZW1jeFdtcE5ORTVxWTNwYWJWcHRDazFxU1RCYVIwVjNXVzFTYTFwcVdUVmFhbEYzV2xSQk1FMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUMFJCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFtZDNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJkNFRsZFNhbHBYV213S1RVZFJlazU2WnpGYWFrMDBUbXBqZWxwdFdtMU5ha2t3V2tkRmQxbHRVbXRhYWxrMVdtcFJkMXBVUVRCTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMDFVVm10Wk1sWnRXbFJDYTAxNll6Uk9WMWw2VDBSWk0wMHlXbTFhYWtsNVRrZFNhRTFIU210YVIxa3lUMWRaTUFwTlIxVjNUa1JCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2swVFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxQUkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSRVV4V2tkT2JBcGFiVlYzV2tSTk0wOUVWbTFOZW1jeVRucE9iVnB0V1hsTmFsSnJXVlJDYVZwSFVtMU9hbXh0VGtSQ2JFMUVVWGRHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDVUa1JuTUU5RVVUVk5SR2MxVERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFpjMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZabEZTTjBGSWEwRmtkMFJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNVpUZEZRVlpCUVVGRlFYZENTVTFGV1VOSlVVTmFLMFpEVkhWNlQycGtZMEkxVTBZeGNpOHdibXRyY0dKcWJYQTFlUXBRTjFjdlNWVmlTMlJSYTA5blVVbG9RVTFTWjBWRVJHTk5jRzlFT0U5RWVFbHVjVlI1YTBGMVZXOXZjRTAzTVZobmRFdDZTVFZIUlZSb2RIWk5RVzlIQ2tORGNVZFRUVFE1UWtGTlJFRXlZMEZOUjFGRFRVWkRaRWRpYjFacGEzVmljVWx0UzAxM1dFaEZXVlFyVlZkVlZHcHhXVTl6Tkc0ek5WQlBabE14T1djS1YwaHdlaTl0Wml0aE0xYzNibWxHUVhGalZHOVZkMGwzVjJ0TGVGRkVNRnBITjNOWVIyeEZPRkJLUzFsalRtWjJZMk5JWVhKUVJtUjZkazF1YTB0bWFRcHlPV0pSWVZsc2RTdElNSEZ0VlVkMGFXSnNPV2N2UjJRS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19",
          "integratedTime": 1772192809,
          "logIndex": 1003522141,
          "logID": "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
        }
      },
      "Issuer": "https://token.actions.githubusercontent.com",
      "Subject": "https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.28",
      "githubWorkflowName": "Release",
      "githubWorkflowRef": "refs/tags/v0.0.28",
      "githubWorkflowRepository": "salrashid123/go-bazel-github-workflow",
      "githubWorkflowSha": "15dcefe0d3785f38673fff224da0bddf69f40e04",
      "githubWorkflowTrigger": "push"
    }
  }
]
```

if you decode the payload

```json
{
  "apiVersion": "0.0.1",
  "kind": "hashedrekord",
  "spec": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "5614df9d0e1fd80e8a4ab2f0af47fcdb4f94b28b4dead1e4b9c78bda31209a8e"
      }
    },
    "signature": {
      "content": "MEYCIQDsZJAMKO1j+mKwoZzQJCWF5DWP91t2L4qW0eu5W/eVTgIhAPzZv0SQjq745Mw2RhmYCRDYRKKhzslDVCC+9jWLiQSk",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMakNDQnJXZ0F3SUJBZ0lVUklOdDhWUURldU9vcFl6UEZpWVN2OExENnFnd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qSTNNVEUwTmpRNFdoY05Nall3TWpJM01URTFOalE0V2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUV5Y242SnNnSSsyRW1QelNGNG1neW9paVJNSFlJSmNmYUhJNjcKODN2SXpCNU5UbEZvQjd6T2t6czRxT08wYktjWlQ5elkyS2Flb3M4N2wvaWFJRTZnL2FPQ0JkUXdnZ1hRTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVUva0VWCmdkMVloSktRSHdrZHhMLzhDS1ZEbzN3d0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJNE1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDZ3hOV1JqWldabE1HUXpOemcxWmpNNE5qY3pabVptCk1qSTBaR0V3WW1Sa1pqWTVaalF3WlRBME1CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5T0RBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNamd3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2d4TldSalpXWmwKTUdRek56ZzFaak00TmpjelptWm1NakkwWkdFd1ltUmtaalk1WmpRd1pUQTBNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b01UVmtZMlZtWlRCa016YzROV1l6T0RZM00yWm1aakl5TkdSaE1HSmtaR1kyT1dZMApNR1V3TkRBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakk0TUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlPREE0QmdvckJnRUVBWU8vTUFFVEJDb01LREUxWkdObApabVV3WkRNM09EVm1NemcyTnpObVptWXlNalJrWVRCaVpHUm1OamxtTkRCbE1EUXdGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl5TkRnME9EUTVNRGc1TDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZc0dDaXNHQVFRQjFua0NCQUlFZlFSN0FIa0Fkd0RkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp5ZTdFQVZBQUFFQXdCSU1FWUNJUUNaK0ZDVHV6T2pkY0I1U0Yxci8wbmtrcGJqbXA1eQpQN1cvSVViS2RRa09nUUloQU1SZ0VERGNNcG9EOE9EeElucVR5a0F1VW9vcE03MVhndEt6STVHRVRodHZNQW9HCkNDcUdTTTQ5QkFNREEyY0FNR1FDTUZDZEdib1Zpa3VicUltS013WEhFWVQrVVdVVGpxWU9zNG4zNVBPZlMxOWcKV0hwei9tZithM1c3bmlGQXFjVG9Vd0l3V2tLeFFEMFpHN3NYR2xFOFBKS1ljTmZ2Y2NIYXJQRmR6dk1ua0tmaQpyOWJRYVlsdStIMHFtVUd0aWJsOWcvR2QKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

the content is the public key

```bash
cat public.crt

-----BEGIN CERTIFICATE-----
MIIHLjCCBrWgAwIBAgIURINt8VQDeuOopYzPFiYSv8LD6qgwCgYIKoZIzj0EAwMw
NzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl
cm1lZGlhdGUwHhcNMjYwMjI3MTE0NjQ4WhcNMjYwMjI3MTE1NjQ4WjAAMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAEycn6JsgI+2EmPzSF4mgyoiiRMHYIJcfaHI67
83vIzB5NTlFoB7zOkzs4qOO0bKcZT9zY2Kaeos87l/iaIE6g/aOCBdQwggXQMA4G
A1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU/kEV
gd1YhJKQHwkdxL/8CKVDo3wwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y
ZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy
My9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs
ZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI4MDkGCisGAQQBg78wAQEEK2h0dHBz
Oi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD
vzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBCgxNWRjZWZlMGQzNzg1ZjM4NjczZmZm
MjI0ZGEwYmRkZjY5ZjQwZTA0MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB
BAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf
BgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yODA7BgorBgEEAYO/MAEIBC0M
K2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK
KwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv
LWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl
LnlhbWxAcmVmcy90YWdzL3YwLjAuMjgwOAYKKwYBBAGDvzABCgQqDCgxNWRjZWZl
MGQzNzg1ZjM4NjczZmZmMjI0ZGEwYmRkZjY5ZjQwZTA0MB0GCisGAQQBg78wAQsE
DwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi
LmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG
AQQBg78wAQ0EKgwoMTVkY2VmZTBkMzc4NWYzODY3M2ZmZjIyNGRhMGJkZGY2OWY0
MGUwNDAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI4MBoGCisGAQQB
g78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0
aHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5
BgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv
Z28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh
c2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yODA4BgorBgEEAYO/MAETBCoMKDE1ZGNl
ZmUwZDM3ODVmMzg2NzNmZmYyMjRkYTBiZGRmNjlmNDBlMDQwFAYKKwYBBAGDvzAB
FAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh
bHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z
LzIyNDg0ODQ5MDg5L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw
gYsGCisGAQQB1nkCBAIEfQR7AHkAdwDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H
inKALynujgAAAZye7EAVAAAEAwBIMEYCIQCZ+FCTuzOjdcB5SF1r/0nkkpbjmp5y
P7W/IUbKdQkOgQIhAMRgEDDcMpoD8ODxInqTykAuUoopM71XgtKzI5GEThtvMAoG
CCqGSM49BAMDA2cAMGQCMFCdGboVikubqImKMwXHEYT+UWUTjqYOs4n35POfS19g
WHpz/mf+a3W7niFAqcToUwIwWkKxQD0ZG7sXGlE8PJKYcNfvccHarPFdzvMnkKfi
r9bQaYlu+H0qmUGtibl9g/Gd
-----END CERTIFICATE-----
```


which includes

```bash
$  openssl x509 -in public.crt -noout -text


Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            44:83:6d:f1:54:03:7a:e3:a8:a5:8c:cf:16:26:12:bf:c2:c3:ea:a8
        Signature Algorithm: ecdsa-with-SHA384
        Issuer: O=sigstore.dev, CN=sigstore-intermediate
        Validity
            Not Before: Feb 27 11:46:48 2026 GMT
            Not After : Feb 27 11:56:48 2026 GMT
        Subject: 
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:c9:c9:fa:26:c8:08:fb:61:26:3f:34:85:e2:68:
                    32:a2:28:91:30:76:08:25:c7:da:1c:8e:bb:f3:7b:
                    c8:cc:1e:4d:4e:51:68:07:bc:ce:93:3b:38:a8:e3:
                    b4:6c:a7:19:4f:dc:d8:d8:a6:9e:a2:cf:3b:97:f8:
                    9a:20:4e:a0:fd
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage: 
                Code Signing
            X509v3 Subject Key Identifier: 
                FE:41:15:81:DD:58:84:92:90:1F:09:1D:C4:BF:FC:08:A5:43:A3:7C
            X509v3 Authority Key Identifier: 
                DF:D3:E9:CF:56:24:11:96:F9:A8:D8:E9:28:55:A2:C6:2E:18:64:3F
            X509v3 Subject Alternative Name: critical
                URI:https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.28
            1.3.6.1.4.1.57264.1.1: 
                https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.2: 
                push
            1.3.6.1.4.1.57264.1.3: 
                15dcefe0d3785f38673fff224da0bddf69f40e04
            1.3.6.1.4.1.57264.1.4: 
                Release
            1.3.6.1.4.1.57264.1.5: 
                salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.6: 
                refs/tags/v0.0.28
            1.3.6.1.4.1.57264.1.8: 
                .+https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.9: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.28
            1.3.6.1.4.1.57264.1.10: 
                .(15dcefe0d3785f38673fff224da0bddf69f40e04
            1.3.6.1.4.1.57264.1.11: 
github-hosted   .
            1.3.6.1.4.1.57264.1.12: 
                .8https://github.com/salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.13: 
                .(15dcefe0d3785f38673fff224da0bddf69f40e04
            1.3.6.1.4.1.57264.1.14: 
                ..refs/tags/v0.0.28
            1.3.6.1.4.1.57264.1.15: 
                .
1150755799
            1.3.6.1.4.1.57264.1.16: 
                ..https://github.com/salrashid123
            1.3.6.1.4.1.57264.1.17: 
                ..11149054
            1.3.6.1.4.1.57264.1.18: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.28
            1.3.6.1.4.1.57264.1.19: 
                .(15dcefe0d3785f38673fff224da0bddf69f40e04
            1.3.6.1.4.1.57264.1.20: 
                ..push
            1.3.6.1.4.1.57264.1.21: 
                .\https://github.com/salrashid123/go-bazel-github-workflow/actions/runs/22484849089/attempts/1
            1.3.6.1.4.1.57264.1.22: 
                ..public
            CT Precertificate SCTs: 
                Signed Certificate Timestamp:
                    Version   : v1 (0x0)
                    Log ID    : DD:3D:30:6A:C6:C7:11:32:63:19:1E:1C:99:67:37:02:
                                A2:4A:5E:B8:DE:3C:AD:FF:87:8A:72:80:2F:29:EE:8E
                    Timestamp : Feb 27 11:46:48.981 2026 GMT
                    Extensions: none
                    Signature : ecdsa-with-SHA256
                                30:46:02:21:00:99:F8:50:93:BB:33:A3:75:C0:79:48:
                                5D:6B:FF:49:E4:92:96:E3:9A:9E:72:3F:B5:BF:21:46:
                                CA:75:09:0E:81:02:21:00:C4:60:10:30:DC:32:9A:03:
                                F0:E0:F1:22:7A:93:CA:40:2E:52:8A:29:33:BD:57:82:
                                D2:B3:23:91:84:4E:1B:6F
    Signature Algorithm: ecdsa-with-SHA384
    Signature Value:
        30:64:02:30:50:9d:19:ba:15:8a:4b:9b:a8:89:8a:33:05:c7:
        11:84:fe:51:65:13:8e:a6:0e:b3:89:f7:e4:f3:9f:4b:5f:60:
        58:7a:73:fe:67:fe:6b:75:bb:9e:21:40:a9:c4:e8:53:02:30:
        5a:42:b1:40:3d:19:1b:bb:17:1a:51:3c:3c:92:98:70:d7:ef:
        71:c1:da:ac:f1:5d:ce:f3:27:90:a7:e2:af:d6:d0:69:89:6e:
        f8:7d:2a:99:41:ad:89:b9:7d:83:f1:9d
```


Now you can use `rekor-cli` to inspect the tlog entry

```bash
$ rekor-cli search --rekor_server https://rekor.sigstore.dev    --sha  5614df9d0e1fd80e8a4ab2f0af47fcdb4f94b28b4dead1e4b9c78bda31209a8e
Found matching entries (listed by UUID):
108e9186e8c5677a9ffc67519139ac94cb3d6919d046a36a79d997d55896eeda9d7063e9bd1c741c
```

for our log entry

```bash
$ rekor-cli get --rekor_server https://rekor.sigstore.dev    --uuid 108e9186e8c5677a9ffc67519139ac94cb3d6919d046a36a79d997d55896eeda9d7063e9bd1c741c 

LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 1003522141
IntegratedTime: 2026-02-27T11:46:49Z
UUID: 108e9186e8c5677a9ffc67519139ac94cb3d6919d046a36a79d997d55896eeda9d7063e9bd1c741c
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "5614df9d0e1fd80e8a4ab2f0af47fcdb4f94b28b4dead1e4b9c78bda31209a8e"
      }
    },
    "signature": {
      "content": "MEYCIQDsZJAMKO1j+mKwoZzQJCWF5DWP91t2L4qW0eu5W/eVTgIhAPzZv0SQjq745Mw2RhmYCRDYRKKhzslDVCC+9jWLiQSk",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMakNDQnJXZ0F3SUJBZ0lVUklOdDhWUURldU9vcFl6UEZpWVN2OExENnFnd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qSTNNVEUwTmpRNFdoY05Nall3TWpJM01URTFOalE0V2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUV5Y242SnNnSSsyRW1QelNGNG1neW9paVJNSFlJSmNmYUhJNjcKODN2SXpCNU5UbEZvQjd6T2t6czRxT08wYktjWlQ5elkyS2Flb3M4N2wvaWFJRTZnL2FPQ0JkUXdnZ1hRTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVUva0VWCmdkMVloSktRSHdrZHhMLzhDS1ZEbzN3d0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJNE1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDZ3hOV1JqWldabE1HUXpOemcxWmpNNE5qY3pabVptCk1qSTBaR0V3WW1Sa1pqWTVaalF3WlRBME1CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5T0RBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNamd3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2d4TldSalpXWmwKTUdRek56ZzFaak00TmpjelptWm1NakkwWkdFd1ltUmtaalk1WmpRd1pUQTBNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b01UVmtZMlZtWlRCa016YzROV1l6T0RZM00yWm1aakl5TkdSaE1HSmtaR1kyT1dZMApNR1V3TkRBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakk0TUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlPREE0QmdvckJnRUVBWU8vTUFFVEJDb01LREUxWkdObApabVV3WkRNM09EVm1NemcyTnpObVptWXlNalJrWVRCaVpHUm1OamxtTkRCbE1EUXdGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl5TkRnME9EUTVNRGc1TDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZc0dDaXNHQVFRQjFua0NCQUlFZlFSN0FIa0Fkd0RkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp5ZTdFQVZBQUFFQXdCSU1FWUNJUUNaK0ZDVHV6T2pkY0I1U0Yxci8wbmtrcGJqbXA1eQpQN1cvSVViS2RRa09nUUloQU1SZ0VERGNNcG9EOE9EeElucVR5a0F1VW9vcE03MVhndEt6STVHRVRodHZNQW9HCkNDcUdTTTQ5QkFNREEyY0FNR1FDTUZDZEdib1Zpa3VicUltS013WEhFWVQrVVdVVGpxWU9zNG4zNVBPZlMxOWcKV0hwei9tZithM1c3bmlGQXFjVG9Vd0l3V2tLeFFEMFpHN3NYR2xFOFBKS1ljTmZ2Y2NIYXJQRmR6dk1ua0tmaQpyOWJRYVlsdStIMHFtVUd0aWJsOWcvR2QKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

you can use crane to inspect the signature too

```bash
$ crane  manifest salrashid123/server_image:sha256-8c0d347ab4b606c93ba087efc5c0e6e58154b052b0eb1d658a313cdaeff93d71.sig | jq '.'


{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 233,
    "digest": "sha256:541151ead71d674686cad365de2608b8fdd62a7a7c063788775c8641d4cdc613"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:5614df9d0e1fd80e8a4ab2f0af47fcdb4f94b28b4dead1e4b9c78bda31209a8e",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEYCIQDsZJAMKO1j+mKwoZzQJCWF5DWP91t2L4qW0eu5W/eVTgIhAPzZv0SQjq745Mw2RhmYCRDYRKKhzslDVCC+9jWLiQSk",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEUCIEMK6jfMz5mP8QP1hehjMOQzZnW7JaN56h/P5LprNF3LAiEAz/upDohXVl+EediGiPEL7mvVacbIGKHlrhVSw0Q4CXw=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI1NjE0ZGY5ZDBlMWZkODBlOGE0YWIyZjBhZjQ3ZmNkYjRmOTRiMjhiNGRlYWQxZTRiOWM3OGJkYTMxMjA5YThlIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FWUNJUURzWkpBTUtPMWorbUt3b1p6UUpDV0Y1RFdQOTF0Mkw0cVcwZXU1Vy9lVlRnSWhBUHpadjBTUWpxNzQ1TXcyUmhtWUNSRFlSS0toenNsRFZDQys5aldMaVFTayIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1ha05EUW5KWFowRjNTVUpCWjBsVlVrbE9kRGhXVVVSbGRVOXZjRmw2VUVacFdWTjJPRXhFTm5GbmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxU1ROTlZFVXdUbXBSTkZkb1kwNU5hbGwzVFdwSk0wMVVSVEZPYWxFMFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVY1WTI0MlNuTm5TU3N5UlcxUWVsTkdORzFuZVc5cGFWSk5TRmxKU21ObVlVaEpOamNLT0ROMlNYcENOVTVVYkVadlFqZDZUMnQ2Y3pSeFQwOHdZa3RqV2xRNWVsa3lTMkZsYjNNNE4yd3ZhV0ZKUlRabkwyRlBRMEprVVhkbloxaFJUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlV2YTBWV0NtZGtNVmxvU2t0UlNIZHJaSGhNTHpoRFMxWkViek4zZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTkUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRFozaE9WMUpxV2xkYWJFMUhVWHBPZW1jeFdtcE5ORTVxWTNwYWJWcHRDazFxU1RCYVIwVjNXVzFTYTFwcVdUVmFhbEYzV2xSQk1FMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUMFJCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFtZDNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJkNFRsZFNhbHBYV213S1RVZFJlazU2WnpGYWFrMDBUbXBqZWxwdFdtMU5ha2t3V2tkRmQxbHRVbXRhYWxrMVdtcFJkMXBVUVRCTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMDFVVm10Wk1sWnRXbFJDYTAxNll6Uk9WMWw2VDBSWk0wMHlXbTFhYWtsNVRrZFNhRTFIU210YVIxa3lUMWRaTUFwTlIxVjNUa1JCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2swVFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxQUkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSRVV4V2tkT2JBcGFiVlYzV2tSTk0wOUVWbTFOZW1jeVRucE9iVnB0V1hsTmFsSnJXVlJDYVZwSFVtMU9hbXh0VGtSQ2JFMUVVWGRHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDVUa1JuTUU5RVVUVk5SR2MxVERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFpjMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZabEZTTjBGSWEwRmtkMFJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNVpUZEZRVlpCUVVGRlFYZENTVTFGV1VOSlVVTmFLMFpEVkhWNlQycGtZMEkxVTBZeGNpOHdibXRyY0dKcWJYQTFlUXBRTjFjdlNWVmlTMlJSYTA5blVVbG9RVTFTWjBWRVJHTk5jRzlFT0U5RWVFbHVjVlI1YTBGMVZXOXZjRTAzTVZobmRFdDZTVFZIUlZSb2RIWk5RVzlIQ2tORGNVZFRUVFE1UWtGTlJFRXlZMEZOUjFGRFRVWkRaRWRpYjFacGEzVmljVWx0UzAxM1dFaEZXVlFyVlZkVlZHcHhXVTl6Tkc0ek5WQlBabE14T1djS1YwaHdlaTl0Wml0aE0xYzNibWxHUVhGalZHOVZkMGwzVjJ0TGVGRkVNRnBITjNOWVIyeEZPRkJLUzFsalRtWjJZMk5JWVhKUVJtUjZkazF1YTB0bWFRcHlPV0pSWVZsc2RTdElNSEZ0VlVkMGFXSnNPV2N2UjJRS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19\",\"integratedTime\":1772192809,\"logIndex\":1003522141,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
        "dev.sigstore.cosign/certificate": "-----BEGIN CERTIFICATE-----\nMIIHLjCCBrWgAwIBAgIURINt8VQDeuOopYzPFiYSv8LD6qgwCgYIKoZIzj0EAwMw\nNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl\ncm1lZGlhdGUwHhcNMjYwMjI3MTE0NjQ4WhcNMjYwMjI3MTE1NjQ4WjAAMFkwEwYH\nKoZIzj0CAQYIKoZIzj0DAQcDQgAEycn6JsgI+2EmPzSF4mgyoiiRMHYIJcfaHI67\n83vIzB5NTlFoB7zOkzs4qOO0bKcZT9zY2Kaeos87l/iaIE6g/aOCBdQwggXQMA4G\nA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU/kEV\ngd1YhJKQHwkdxL/8CKVDo3wwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y\nZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy\nMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs\nZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI4MDkGCisGAQQBg78wAQEEK2h0dHBz\nOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD\nvzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBCgxNWRjZWZlMGQzNzg1ZjM4NjczZmZm\nMjI0ZGEwYmRkZjY5ZjQwZTA0MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB\nBAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf\nBgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yODA7BgorBgEEAYO/MAEIBC0M\nK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK\nKwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv\nLWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl\nLnlhbWxAcmVmcy90YWdzL3YwLjAuMjgwOAYKKwYBBAGDvzABCgQqDCgxNWRjZWZl\nMGQzNzg1ZjM4NjczZmZmMjI0ZGEwYmRkZjY5ZjQwZTA0MB0GCisGAQQBg78wAQsE\nDwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi\nLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG\nAQQBg78wAQ0EKgwoMTVkY2VmZTBkMzc4NWYzODY3M2ZmZjIyNGRhMGJkZGY2OWY0\nMGUwNDAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI4MBoGCisGAQQB\ng78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0\naHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5\nBgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv\nZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh\nc2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yODA4BgorBgEEAYO/MAETBCoMKDE1ZGNl\nZmUwZDM3ODVmMzg2NzNmZmYyMjRkYTBiZGRmNjlmNDBlMDQwFAYKKwYBBAGDvzAB\nFAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh\nbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z\nLzIyNDg0ODQ5MDg5L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw\ngYsGCisGAQQB1nkCBAIEfQR7AHkAdwDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H\ninKALynujgAAAZye7EAVAAAEAwBIMEYCIQCZ+FCTuzOjdcB5SF1r/0nkkpbjmp5y\nP7W/IUbKdQkOgQIhAMRgEDDcMpoD8ODxInqTykAuUoopM71XgtKzI5GEThtvMAoG\nCCqGSM49BAMDA2cAMGQCMFCdGboVikubqImKMwXHEYT+UWUTjqYOs4n35POfS19g\nWHpz/mf+a3W7niFAqcToUwIwWkKxQD0ZG7sXGlE8PJKYcNfvccHarPFdzvMnkKfi\nr9bQaYlu+H0qmUGtibl9g/Gd\n-----END CERTIFICATE-----\n",
        "dev.sigstore.cosign/chain": "-----BEGIN CERTIFICATE-----\nMIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRkHrn/UnfaQwCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMjA0MTMyMDA2MTVaFw0zMTEwMDUxMzU2NThaMDcxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjEeMBwGA1UEAxMVc2lnc3RvcmUtaW50ZXJtZWRpYXRlMHYwEAYHKoZIzj0C\nAQYFK4EEACIDYgAE8RVS/ysH+NOvuDZyPIZtilgUF9NlarYpAd9HP1vBBH1U5CV7\n7LSS7s0ZiH4nE7Hv7ptS6LvvR/STk798LVgMzLlJ4HeIfF3tHSaexLcYpSASr1kS\n0N/RgBJz/9jWCiXno3sweTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYB\nBQUHAwMwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU39Ppz1YkEZb5qNjp\nKFWixi4YZD8wHwYDVR0jBBgwFoAUWMAeX5FFpWapesyQoZMi0CrFxfowCgYIKoZI\nzj0EAwMDZwAwZAIwPCsQK4DYiZYDPIaDi5HFKnfxXx6ASSVmERfsynYBiX2X6SJR\nnZU84/9DZdnFvvxmAjBOt6QpBlc4J/0DxvkTCqpclvziL6BCCPnjdlIB3Pu3BxsP\nmygUY7Ii2zbdCdliiow=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIB9zCCAXygAwIBAgIUALZNAPFdxHPwjeDloDwyYChAO/4wCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMTEwMDcxMzU2NTlaFw0zMTEwMDUxMzU2NThaMCoxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjERMA8GA1UEAxMIc2lnc3RvcmUwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT7\nXeFT4rb3PQGwS4IajtLk3/OlnpgangaBclYpsYBr5i+4ynB07ceb3LP0OIOZdxex\nX69c5iVuyJRQ+Hz05yi+UF3uBWAlHpiS5sh0+H2GHE7SXrk1EC5m1Tr19L9gg92j\nYzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRY\nwB5fkUWlZql6zJChkyLQKsXF+jAfBgNVHSMEGDAWgBRYwB5fkUWlZql6zJChkyLQ\nKsXF+jAKBggqhkjOPQQDAwNpADBmAjEAj1nHeXZp+13NWBNa+EDsDP8G1WWg1tCM\nWP/WHPqpaVo0jhsweNFZgSs0eE7wYI4qAjEA2WB9ot98sIkoF3vZYdd3/VtWB5b9\nTNMea7Ix/stJ5TfcLLeABLE4BNJOsQ4vnBHJ\n-----END CERTIFICATE-----"
      }
    }
  ]
}
```


#### Verify binary

THe current binary uses github workflows to sign.  If you want to use local keys provided in this repo, edit `sign_binary` genrule

The github workflow also uses bazel to sign each image usign [cosign sign-blob](https://edu.chainguard.dev/open-source/sigstore/cosign/how-to-sign-blobs-with-cosign/)

If you want to use bazel manually with local certs:

```bash
### generate binary
$ bazelisk build  app:build_binary 
  bazel-bin/app/server_linux_amd64_bin
  bazel-bin/app/server_linux_arm64_bin

### sign with key
#### to use this edit app/BUILD.bazel and edit sign_binary rule as commented there
$ bazelisk build  app:sign_binary
  bazel-bin/app/server_linux_amd64.sig
  bazel-bin/app/server_linux_arm64.sig

### a local signature may look like

$ cat bazel-bin/app/server_linux_amd64.sig | jq '.'

{
  "mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json",
  "verificationMaterial": {
    "publicKey": {
      "hint": "cka9QDcaiM9PNxTe4aOOEWcFiaRm0y59r4YcwitHqBc="
    }
  },
  "messageSignature": {
    "messageDigest": {
      "algorithm": "SHA2_256",
      "digest": "bgCKnEjVBu0hi9IsiBjOGLVfQaahv+x2UXp2FMV6LNw="
    },
    "signature": "ELLlVCZ9ZYA6W/IOzdCBlt3pJwenXsdZNvNsgsHb2QPQF+RA8VSDVppP5qhy/3RMh1kpH/rSpJgdBJX9sfl5DEIiNi+9lKIstXn4ErLyZbwBN3bk/LVHAzxKs71iH+s0GMDyYWiEsEnUPaGGxEui7FnEHNElpxcnbeBj7UttfNSsgxxmpU4XH+VwpXkiqOON8tWJ2R0g4RvEE0HvA/CZd4XJkI6v2lVqIkXit5gENMzuUGkH74m9rDLNGz2GXQ7DNYc6w6Ca3paq1Vm8ZDvZ7zwh7VZAg9fbRWleVwdP0HeyF3SfSNcjsVebhfzkJ9xssF2CSkTMg27Yo+tg2fuA7Q=="
  }
}


### which can be verified locally with the local signing key 
export sig=`cat bazel-bin/app/server_linux_amd64.sig | jq -r '.messageSignature.signature'`

## use --insecure-ignore-tlog=true since we didn't upload the local signed file
cosign verify-blob --insecure-ignore-tlog=true --key certs/import-cosign.pub --signature $sig bazel-bin/app/server_linux_amd64_bin
```

To verify the live binary,

```bash
wget https://github.com/salrashid123/go-bazel-github-workflow/releases/download/v0.0.28/server_linux_amd64.sig
wget https://github.com/salrashid123/go-bazel-github-workflow/releases/download/v0.0.28/server_linux_amd64_bin

cat server_linux_amd64.sig | jq '.'

{
  "mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json",
  "verificationMaterial": {
    "certificate": {
      "rawBytes": "MIIHLjCCBrOgAwIBAgIUEEXG2sm/TS3a4W9gf0XidSiHxbUwCgYIKoZIzj0EAwMwNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRlcm1lZGlhdGUwHhcNMjYwMjI3MTE0NjIzWhcNMjYwMjI3MTE1NjIzWjAAMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkRspbWS59zzN5Wdpb0xmMYZacbjQkLQwRk3xu1vwoLn4BSJ9WkD44tCSiXtg+HO5qkKZeIydcBZHdT03O8nYvKOCBdIwggXOMA4GA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQUKf1v7Erjg4NbkqmN2pNfcBVatCgwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4YZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVsZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI4MDkGCisGAQQBg78wAQEEK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGDvzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBCgxNWRjZWZlMGQzNzg1ZjM4NjczZmZmMjI0ZGEwYmRkZjY5ZjQwZTA0MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYBBAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAfBgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yODA7BgorBgEEAYO/MAEIBC0MK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYKKwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNlLnlhbWxAcmVmcy90YWdzL3YwLjAuMjgwOAYKKwYBBAGDvzABCgQqDCgxNWRjZWZlMGQzNzg1ZjM4NjczZmZmMjI0ZGEwYmRkZjY5ZjQwZTA0MB0GCisGAQQBg78wAQsEDwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisGAQQBg78wAQ0EKgwoMTVkY2VmZTBkMzc4NWYzODY3M2ZmZjIyNGRhMGJkZGY2OWY0MGUwNDAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI4MBoGCisGAQQBg78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5BgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVhc2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yODA4BgorBgEEAYO/MAETBCoMKDE1ZGNlZmUwZDM3ODVmMzg2NzNmZmYyMjRkYTBiZGRmNjlmNDBlMDQwFAYKKwYBBAGDvzABFAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5zLzIyNDg0ODQ5MDg5L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMwgYkGCisGAQQB1nkCBAIEewR5AHcAdQDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+HinKALynujgAAAZye69sYAAAEAwBGMEQCICTVvgwytzUuugQAXMapXEztMTOhTBlFxkwKXBMd6a1gAiBoxpaQ8x8IxGyFpjQyIxPhde8GAvj1uA2XUFWxoLj+bDAKBggqhkjOPQQDAwNpADBmAjEAgu6r1Oi8fJMeMP6fu/DeS1NSsTZPyrYOUIm75FfEvcmIZp5dr4XlTUeSzQjXGaZmAjEAjLt21Yzwxjzh8dFRDGr5TNeoYFnLpIevN4MW8n2muzb04nN2NtHFgfDp3mMVJgKU"
    },
    "tlogEntries": [
      {
        "logIndex": "1003521593",
        "logId": {
          "keyId": "wNI9atQGlz+VWfO6LRygH4QUfY/8W4RFwiT5i5WRgB0="
        },
        "kindVersion": {
          "kind": "hashedrekord",
          "version": "0.0.1"
        },
        "integratedTime": "1772192783",
        "inclusionPromise": {
          "signedEntryTimestamp": "MEYCIQDIUI2QKv032sMzGorzUGRErRAwYdq3RoPqN8SbLSIzIwIhAKGwpfNLm4DGmXx6hzF3O15jx40GyxujPH+4bSx6160d"
        },
        "inclusionProof": {
          "logIndex": "881617331",
          "rootHash": "xjg2ORwyFVGoiwfxdkZUIOb+zqLJz59LnQXu4Ys3rJQ=",
          "treeSize": "881617336",
          "hashes": [
            "fKfNtCR0OrKpEJ57mrcKXXE2kCAjG1pnfI4aHLDRfqw=",
            "g+c0O5rdKgIHzoRDwIiud6YB15BCl81QJVQ4SkTFjyI=",
            "1QCiF5iv4BTQDDMDvKFOMe8okH9XrDjXDqaQE06Xmdo=",
            "B41hcXkCJgzG0MOxoHk48ORN02Sfen5VdWWzHBi/iaM=",
            "ZjuSSe2uTPBAGGUax9hZMVrDSkCwidKfye9WQbqaUXo=",
            "raozs+Mt7u9RW1wZ2PNMXEKnFYLsh99Jms+NgdCtDu0=",
            "R+VBiPtkmXrotMuYaZJNKLnqUwZarwE2mtSFwLiPB74=",
            "0AG2dLCqJWJRwFqgHf+jKbX2Kddu+QcS1yha7CZwoKo=",
            "VO8kNVq/e4sOmA0lDpZLud5D6Y09Y3hEpjRIyvXdoD4=",
            "3gfgI4xsS+fLKeK13IhfDzRLgurmZSoZIf5eR7CxzjA=",
            "67+lQA3FSRoR69ZInFdOk9ql7TugFmLM3yStCFec+MU=",
            "bn6WxEzu4yJ8EbkTztOYLXzIJvYZc5WiytgfRQvLuiw=",
            "mjr+JtvJ1BwYlJvqagR3tMH25XuTBdYkgN1yMnmeCCs=",
            "ZleKYeRKwUF3HP3HO0kxHMVeJgY3N/euGinVhlVWaq0=",
            "fLAvE46NqCVV86EpB2pKkwJlFjjFk7ntX3lC+PiZuIo=",
            "T4DqWD42hAtN+vX8jKCWqoC4meE4JekI9LxYGCcPy1M="
          ],
          "checkpoint": {
            "envelope": "rekor.sigstore.dev - 1193050959916656506\n881617336\nxjg2ORwyFVGoiwfxdkZUIOb+zqLJz59LnQXu4Ys3rJQ=\n\n— rekor.sigstore.dev wNI9ajBFAiAJ7Nu7B1cK2si+fyQAlzbiWwsG3ZD+lEyY1Uk1nCgMnQIhAPp51nF7KOl6c72IxU9km9mn6mFIJ1Bl5OWS0YDMeCKX\n"
          }
        },
        "canonicalizedBody": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI2ZTAwOGE5YzQ4ZDUwNmVkMjE4YmQyMmM4ODE4Y2UxOGI1NWY0MWE2YTFiZmVjNzY1MTdhNzYxNGM1N2EyY2RjIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRC9xZzV6QU9hbnY2UGsvcDVEVFFzdDh1QW91QUJCd1B6RUttYTRPeGxGS0FpRUFxMjBUZW4zVWVKempVYnlYZll1NlpzbUZZWFJUS1BLWmw1Vzd5N0Z2ME9NPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1ha05EUW5KUFowRjNTVUpCWjBsVlJVVllSekp6YlM5VVV6TmhORmM1WjJZd1dHbGtVMmxJZUdKVmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxU1ROTlZFVXdUbXBKZWxkb1kwNU5hbGwzVFdwSk0wMVVSVEZPYWtsNlYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVZyVW5Od1lsZFROVGw2ZWs0MVYyUndZakI0YlUxWldtRmpZbXBSYTB4UmQxSnJNM2dLZFRGMmQyOU1ialJDVTBvNVYydEVORFIwUTFOcFdIUm5LMGhQTlhGclMxcGxTWGxrWTBKYVNHUlVNRE5QT0c1WmRrdFBRMEprU1hkbloxaFBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlZMWmpGMkNqZEZjbXBuTkU1aWEzRnRUakp3VG1aalFsWmhkRU5uZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTkUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRFozaE9WMUpxV2xkYWJFMUhVWHBPZW1jeFdtcE5ORTVxWTNwYWJWcHRDazFxU1RCYVIwVjNXVzFTYTFwcVdUVmFhbEYzV2xSQk1FMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUMFJCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFtZDNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJkNFRsZFNhbHBYV213S1RVZFJlazU2WnpGYWFrMDBUbXBqZWxwdFdtMU5ha2t3V2tkRmQxbHRVbXRhYWxrMVdtcFJkMXBVUVRCTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMDFVVm10Wk1sWnRXbFJDYTAxNll6Uk9WMWw2VDBSWk0wMHlXbTFhYWtsNVRrZFNhRTFIU210YVIxa3lUMWRaTUFwTlIxVjNUa1JCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2swVFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxQUkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSRVV4V2tkT2JBcGFiVlYzV2tSTk0wOUVWbTFOZW1jeVRucE9iVnB0V1hsTmFsSnJXVlJDYVZwSFVtMU9hbXh0VGtSQ2JFMUVVWGRHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDVUa1JuTUU5RVVUVk5SR2MxVERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFphMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNVpUWTVjMWxCUVVGRlFYZENSMDFGVVVOSlExUldkbWQzZVhSNlZYVjFaMUZCV0UxaGNGaEZlblJOVkU5b1ZFSnNSZ3A0YTNkTFdFSk5aRFpoTVdkQmFVSnZlSEJoVVRoNE9FbDRSM2xHY0dwUmVVbDRVR2hrWlRoSFFYWnFNWFZCTWxoVlJsZDRiMHhxSzJKRVFVdENaMmR4Q21ocmFrOVFVVkZFUVhkT2NFRkVRbTFCYWtWQlozVTJjakZQYVRobVNrMWxUVkEyWm5VdlJHVlRNVTVUYzFSYVVIbHlXVTlWU1cwM05VWm1SWFpqYlVrS1duQTFaSEkwV0d4VVZXVlRlbEZxV0VkaFdtMUJha1ZCYWt4ME1qRlplbmQ0YW5wb09HUkdVa1JIY2pWVVRtVnZXVVp1VEhCSlpYWk9ORTFYT0c0eWJRcDFlbUl3Tkc1T01rNTBTRVpuWmtSd00yMU5Wa3BuUzFVS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19"
      }
    ],
    "timestampVerificationData": {
      "rfc3161Timestamps": [
        {
          "signedTimestamp": "MIICyDADAgEAMIICvwYJKoZIhvcNAQcCoIICsDCCAqwCAQMxDTALBglghkgBZQMEAgEwgbcGCyqGSIb3DQEJEAEEoIGnBIGkMIGhAgEBBgkrBgEEAYO/MAIwMTANBglghkgBZQMEAgEFAAQg5ETmB3ckxxRhUDPLEEbWkfivG3HTnPIG9izTj3ws4HsCFHdlUTjobTTxgpjZaiE3c8Cg/HJmGA8yMDI2MDIyNzExNDYyM1owAwIBAaAypDAwLjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MRUwEwYDVQQDEwxzaWdzdG9yZS10c2GgADGCAdowggHWAgEBMFEwOTEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MSAwHgYDVQQDExdzaWdzdG9yZS10c2Etc2VsZnNpZ25lZAIUOhNULwyQYe68wUMvy4qOiyojiwwwCwYJYIZIAWUDBAIBoIH8MBoGCSqGSIb3DQEJAzENBgsqhkiG9w0BCRABBDAcBgkqhkiG9w0BCQUxDxcNMjYwMjI3MTE0NjIzWjAvBgkqhkiG9w0BCQQxIgQg812TAafk7N67e+hbtZUI45lW3xI3dhXP4iauaMCO1QowgY4GCyqGSIb3DQEJEAIvMX8wfTB7MHkEIIX5J7wHq2LKw7RDVsEO/IGyxog/2nq55thw2dE6zQW3MFUwPaQ7MDkxFTATBgNVBAoTDHNpZ3N0b3JlLmRldjEgMB4GA1UEAxMXc2lnc3RvcmUtdHNhLXNlbGZzaWduZWQCFDoTVC8MkGHuvMFDL8uKjosqI4sMMAoGCCqGSM49BAMCBGYwZAIwXcCjsZfFpvogR1Phv+nI0dHrcl9f4gNv9PfNyVbKXS8Wn0VuzPbja8ZejwccDM4+AjAul6AJcIlu5N4JLemrSe+zOwLH6gg2BIXxF0CBhVrVvUlpRQT5+C/I9dp78DnWFe8="
        }
      ]
    }
  },
  "messageSignature": {
    "messageDigest": {
      "algorithm": "SHA2_256",
      "digest": "bgCKnEjVBu0hi9IsiBjOGLVfQaahv+x2UXp2FMV6LNw="
    },
    "signature": "MEUCID/qg5zAOanv6Pk/p5DTQst8uAouABBwPzEKma4OxlFKAiEAq20Ten3UeJzjUbyXfYu6ZsmFYXRTKPKZl5W7y7Fv0OM="
  }
}


### which can be verified locally with the local signing key 
export sig=`cat server_linux_amd64.sig | jq -r '.messageSignature.signature'`

## use --insecure-ignore-tlog=true since we didn't upload the local signed file
$ cosign verify-blob --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp="https://github.com.*" \
    --bundle server_linux_amd64.sig server_linux_amd64_bin --verbose
Verified OK

```

