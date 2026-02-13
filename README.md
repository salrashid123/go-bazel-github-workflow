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

export IMAGE="docker.io/salrashid123/server_image:server@sha256:bbfdc97be6fbabb8be57c346c47816c8a95904b14448270cd0585a822c33961a"

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
export IMAGE="docker.io/salrashid123/server_image:server@sha256:bbfdc97be6fbabb8be57c346c47816c8a95904b14448270cd0585a822c33961a"

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

export TAG=v0.0.25
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

Verification for index.docker.io/salrashid123/server_image@sha256:bbfdc97be6fbabb8be57c346c47816c8a95904b14448270cd0585a822c33961a --
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
        "docker-manifest-digest": "sha256:bbfdc97be6fbabb8be57c346c47816c8a95904b14448270cd0585a822c33961a"
      },
      "type": "cosign container image signature"
    },
    "optional": {
      "1.3.6.1.4.1.57264.1.1": "https://token.actions.githubusercontent.com",
      "1.3.6.1.4.1.57264.1.2": "push",
      "1.3.6.1.4.1.57264.1.3": "bdacfdd2c8b922caca665b30be54bf52bec33e77",
      "1.3.6.1.4.1.57264.1.4": "Release",
      "1.3.6.1.4.1.57264.1.5": "salrashid123/go-bazel-github-workflow",
      "1.3.6.1.4.1.57264.1.6": "refs/tags/v0.0.25",
      "Bundle": {
        "SignedEntryTimestamp": "MEUCICvDQTw6TSI3GgrhqzH+AOy0A2k5mMtzoDleadAywHCNAiEAp2DIhQpZQ4Y1vWdB5eRrJZTtA8IubLOLgZM/SWtKqGA=",
        "Payload": {
          "body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJkYTg2Nzk0Y2VhNWE3MmEwY2EyMTZjYTljN2FkOWIzOGIxYzUwZTU2MDRhNzE3NjJiY2FjNThiYWIwNTljYmU4In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRnk1T1BoQVphMkFHR3RGOUkrZlVYUDZ2czUrRFVXZ25tTGl5Q2lkdXhPZ0FpRUF2b0ZrcGdubjhrVFZRSEhEZGtvK2ZYRGhlOHE1NTlnNXJtZ2k2WTkxeStjPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1WRU5EUW5KUFowRjNTVUpCWjBsVlQzQnlSMXB1UlZaeVNHcGtNSFpDVHpOMlZHNHhaekY0WlhkVmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUlhwTlZGVjNUMVJOTUZkb1kwNU5hbGwzVFdwRmVrMVVWWGhQVkUwd1YycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVZTTm1SMWFGSnRla2hDU0dkclJEUlFibmhVT1VzMlZGSnhRVEZ0YTJkNk5rNWpOa1FLUXpremMyVnViVTV6ZGtkdE5GZHJja1EzTjA1c09Vb3JNa2Q1VHl0NWQzQkxXVVpuTjFsYVVVTklUMlJCWTNSTk1VdFBRMEprU1hkbloxaFBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlUzYm5CRkNuWTJVMGxhYW14bmVtaHhiWEV4ZVdReWJXRkZSVXMwZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTVUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRGFHbGFSMFpxV20xU2EwMXRUVFJaYW10NVRXMU9hRmt5UlRKT2FsWnBDazE2UW1sYVZGVXdXVzFaTVUxdFNteFplazE2V2xSak0wMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUbFJCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFsVjNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJocFdrZEdhbHB0VW1zS1RXMU5ORmxxYTNsTmJVNW9XVEpGTWs1cVZtbE5la0pwV2xSVk1GbHRXVEZOYlVwc1dYcE5lbHBVWXpOTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMWx0VW1oWk1scHJXa1JLYWs5SFNUVk5ha3BxV1ZkT2FFNXFXVEZaYWsxM1dXMVZNVTVIU20xT1ZFcHBXbGROZWdwTk1sVXpUbnBCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2t4VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxPVkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSMHByV1ZkT2JRcGFSMUY1V1hwb2FVOVVTWGxaTWtacVdWUlpNazVYU1hwTlIwcHNUbFJTYVZwcVZYbFpiVlpxVFhwT2JFNTZZM2RHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDRUMVJyZUU1cWF6Rk9lbFV5VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFphMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNFdHcE9iMmhCUVVGRlFYZENSMDFGVVVOSlFtSmlkMEZvWm5sQ1FrRjFVVU15YzA5MFpuTk1aVlozVmxKYWRHWnRSd3B6WVRaeFdrVjFlSEpSTVhsQmFVSnBaVVl5ZVhFMFVsaDVZV2R1Y21sRlFUWjNPRXN2TldoUlJrZENRbk00Ym5Ka00wZGFXSEJaTlZoNlFVdENaMmR4Q21ocmFrOVFVVkZFUVhkT2IwRkVRbXhCYWtKeVoxTm9TbWM1T0V4cGEzWjZLeXRSWTBKbGVWSlpSelZJT0dGRFYwNWtUazltT0hsbmQxSnljMjl0VFVFS1ppOVlka1pRZGtoeVMwaFRkR3h1U0dwUVJVTk5VVU14WWtoU2JFaFNZVmxQVkM5VlZsUmhOMHAzWVhoNFpEbFdUbk4wYW5rcmQwNWxVSFV6Yml0Wll3cElkbHAxZUVGemVHdG1SRXBTTTFOTFpDOWpNSEpWT0QwS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19",
          "integratedTime": 1770995375,
          "logIndex": 947581057,
          "logID": "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
        }
      },
      "Issuer": "https://token.actions.githubusercontent.com",
      "Subject": "https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.25",
      "githubWorkflowName": "Release",
      "githubWorkflowRef": "refs/tags/v0.0.25",
      "githubWorkflowRepository": "salrashid123/go-bazel-github-workflow",
      "githubWorkflowSha": "bdacfdd2c8b922caca665b30be54bf52bec33e77",
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
        "value": "da86794cea5a72a0ca216ca9c7ad9b38b1c50e5604a71762bcac58bab059cbe8"
      }
    },
    "signature": {
      "content": "MEUCIFy5OPhAZa2AGGtF9I+fUXP6vs5+DUWgnmLiyCiduxOgAiEAvoFkpgnn8kTVQHHDdko+fXDhe8q559g5rmgi6Y91y+c=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMVENDQnJPZ0F3SUJBZ0lVT3ByR1puRVZySGpkMHZCTzN2VG4xZzF4ZXdVd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qRXpNVFV3T1RNMFdoY05Nall3TWpFek1UVXhPVE0wV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVSNmR1aFJtekhCSGdrRDRQbnhUOUs2VFJxQTFta2d6Nk5jNkQKQzkzc2VubU5zdkdtNFdrckQ3N05sOUorMkd5Tyt5d3BLWUZnN1laUUNIT2RBY3RNMUtPQ0JkSXdnZ1hPTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVU3bnBFCnY2U0laamxnemhxbXExeWQybWFFRUs0d0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJMU1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDaGlaR0ZqWm1Sa01tTTRZamt5TW1OaFkyRTJOalZpCk16QmlaVFUwWW1ZMU1tSmxZek16WlRjM01CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5TlRBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNalV3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2hpWkdGalptUmsKTW1NNFlqa3lNbU5oWTJFMk5qVmlNekJpWlRVMFltWTFNbUpsWXpNelpUYzNNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b1ltUmhZMlprWkRKak9HSTVNakpqWVdOaE5qWTFZak13WW1VMU5HSm1OVEppWldNegpNMlUzTnpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakkxTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlOVEE0QmdvckJnRUVBWU8vTUFFVEJDb01LR0prWVdObQpaR1F5WXpoaU9USXlZMkZqWVRZMk5XSXpNR0psTlRSaVpqVXlZbVZqTXpObE56Y3dGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl4T1RreE5qazFOelUyTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZa0dDaXNHQVFRQjFua0NCQUlFZXdSNUFIY0FkUURkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp4WGpOb2hBQUFFQXdCR01FUUNJQmJid0FoZnlCQkF1UUMyc090ZnNMZVZ3VlJadGZtRwpzYTZxWkV1eHJRMXlBaUJpZUYyeXE0Ulh5YWducmlFQTZ3OEsvNWhRRkdCQnM4bnJkM0daWHBZNVh6QUtCZ2dxCmhrak9QUVFEQXdOb0FEQmxBakJyZ1NoSmc5OExpa3Z6KytRY0JleVJZRzVIOGFDV05kTk9mOHlnd1Jyc29tTUEKZi9YdkZQdkhyS0hTdGxuSGpQRUNNUUMxYkhSbEhSYVlPVC9VVlRhN0p3YXh4ZDlWTnN0ankrd05lUHUzbitZYwpIdlp1eEFzeGtmREpSM1NLZC9jMHJVOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

the content is the public key

```bash
cat public.crt

-----BEGIN CERTIFICATE-----
MIIHLTCCBrOgAwIBAgIUOprGZnEVrHjd0vBO3vTn1g1xewUwCgYIKoZIzj0EAwMw
NzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl
cm1lZGlhdGUwHhcNMjYwMjEzMTUwOTM0WhcNMjYwMjEzMTUxOTM0WjAAMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAER6duhRmzHBHgkD4PnxT9K6TRqA1mkgz6Nc6D
C93senmNsvGm4WkrD77Nl9J+2GyO+ywpKYFg7YZQCHOdActM1KOCBdIwggXOMA4G
A1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU7npE
v6SIZjlgzhqmq1yd2maEEK4wHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y
ZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy
My9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs
ZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI1MDkGCisGAQQBg78wAQEEK2h0dHBz
Oi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD
vzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBChiZGFjZmRkMmM4YjkyMmNhY2E2NjVi
MzBiZTU0YmY1MmJlYzMzZTc3MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB
BAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf
BgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yNTA7BgorBgEEAYO/MAEIBC0M
K2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK
KwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv
LWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl
LnlhbWxAcmVmcy90YWdzL3YwLjAuMjUwOAYKKwYBBAGDvzABCgQqDChiZGFjZmRk
MmM4YjkyMmNhY2E2NjViMzBiZTU0YmY1MmJlYzMzZTc3MB0GCisGAQQBg78wAQsE
DwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi
LmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG
AQQBg78wAQ0EKgwoYmRhY2ZkZDJjOGI5MjJjYWNhNjY1YjMwYmU1NGJmNTJiZWMz
M2U3NzAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI1MBoGCisGAQQB
g78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0
aHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5
BgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv
Z28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh
c2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yNTA4BgorBgEEAYO/MAETBCoMKGJkYWNm
ZGQyYzhiOTIyY2FjYTY2NWIzMGJlNTRiZjUyYmVjMzNlNzcwFAYKKwYBBAGDvzAB
FAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh
bHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z
LzIxOTkxNjk1NzU2L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw
gYkGCisGAQQB1nkCBAIEewR5AHcAdQDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H
inKALynujgAAAZxXjNohAAAEAwBGMEQCIBbbwAhfyBBAuQC2sOtfsLeVwVRZtfmG
sa6qZEuxrQ1yAiBieF2yq4RXyagnriEA6w8K/5hQFGBBs8nrd3GZXpY5XzAKBggq
hkjOPQQDAwNoADBlAjBrgShJg98Likvz++QcBeyRYG5H8aCWNdNOf8ygwRrsomMA
f/XvFPvHrKHStlnHjPECMQC1bHRlHRaYOT/UVTa7Jwaxxd9VNstjy+wNePu3n+Yc
HvZuxAsxkfDJR3SKd/c0rU8=
-----END CERTIFICATE-----
```


which includes

```bash
$  openssl x509 -in public.crt -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            3a:9a:c6:66:71:15:ac:78:dd:d2:f0:4e:de:f4:e7:d6:0d:71:7b:05
        Signature Algorithm: ecdsa-with-SHA384
        Issuer: O=sigstore.dev, CN=sigstore-intermediate
        Validity
            Not Before: Feb 13 15:09:34 2026 GMT
            Not After : Feb 13 15:19:34 2026 GMT
        Subject: 
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:47:a7:6e:85:19:b3:1c:11:e0:90:3e:0f:9f:14:
                    fd:2b:a4:d1:a8:0d:66:92:0c:fa:35:ce:83:0b:dd:
                    ec:7a:79:8d:b2:f1:a6:e1:69:2b:0f:be:cd:97:d2:
                    7e:d8:6c:8e:fb:2c:29:29:81:60:ed:86:50:08:73:
                    9d:01:cb:4c:d4
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage: 
                Code Signing
            X509v3 Subject Key Identifier: 
                EE:7A:44:BF:A4:88:66:39:60:CE:1A:A6:AB:5C:9D:DA:66:84:10:AE
            X509v3 Authority Key Identifier: 
                DF:D3:E9:CF:56:24:11:96:F9:A8:D8:E9:28:55:A2:C6:2E:18:64:3F
            X509v3 Subject Alternative Name: critical
                URI:https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.25
            1.3.6.1.4.1.57264.1.1: 
                https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.2: 
                push
            1.3.6.1.4.1.57264.1.3: 
                bdacfdd2c8b922caca665b30be54bf52bec33e77
            1.3.6.1.4.1.57264.1.4: 
                Release
            1.3.6.1.4.1.57264.1.5: 
                salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.6: 
                refs/tags/v0.0.25
            1.3.6.1.4.1.57264.1.8: 
                .+https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.9: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.25
            1.3.6.1.4.1.57264.1.10: 
                .(bdacfdd2c8b922caca665b30be54bf52bec33e77
            1.3.6.1.4.1.57264.1.11: 
github-hosted   .
            1.3.6.1.4.1.57264.1.12: 
                .8https://github.com/salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.13: 
                .(bdacfdd2c8b922caca665b30be54bf52bec33e77
            1.3.6.1.4.1.57264.1.14: 
                ..refs/tags/v0.0.25
            1.3.6.1.4.1.57264.1.15: 
                .
1150755799
            1.3.6.1.4.1.57264.1.16: 
                ..https://github.com/salrashid123
            1.3.6.1.4.1.57264.1.17: 
                ..11149054
            1.3.6.1.4.1.57264.1.18: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.25
            1.3.6.1.4.1.57264.1.19: 
                .(bdacfdd2c8b922caca665b30be54bf52bec33e77
            1.3.6.1.4.1.57264.1.20: 
                ..push
            1.3.6.1.4.1.57264.1.21: 
                .\https://github.com/salrashid123/go-bazel-github-workflow/actions/runs/21991695756/attempts/1
            1.3.6.1.4.1.57264.1.22: 
                ..public
            CT Precertificate SCTs: 
                Signed Certificate Timestamp:
                    Version   : v1 (0x0)
                    Log ID    : DD:3D:30:6A:C6:C7:11:32:63:19:1E:1C:99:67:37:02:
                                A2:4A:5E:B8:DE:3C:AD:FF:87:8A:72:80:2F:29:EE:8E
                    Timestamp : Feb 13 15:09:34.625 2026 GMT
                    Extensions: none
                    Signature : ecdsa-with-SHA256
                                30:44:02:20:16:DB:C0:08:5F:C8:10:40:B9:00:B6:B0:
                                EB:5F:B0:B7:95:C1:54:59:B5:F9:86:B1:AE:AA:64:4B:
                                B1:AD:0D:72:02:20:62:78:5D:B2:AB:84:57:C9:A8:27:
                                AE:21:00:EB:0F:0A:FF:98:50:14:60:41:B3:C9:EB:77:
                                71:99:5E:96:39:5F
    Signature Algorithm: ecdsa-with-SHA384
    Signature Value:
        30:65:02:30:6b:81:28:49:83:df:0b:8a:4b:f3:fb:e4:1c:05:
        ec:91:60:6e:47:f1:a0:96:35:d3:4e:7f:cc:a0:c1:1a:ec:a2:
        63:00:7f:f5:ef:14:fb:c7:ac:a1:d2:b6:59:c7:8c:f1:02:31:
        00:b5:6c:74:65:1d:16:98:39:3f:d4:55:36:bb:27:06:b1:c5:
        df:55:36:cb:63:cb:ec:0d:78:fb:b7:9f:e6:1c:1e:f6:6e:c4:
        0b:31:91:f0:c9:47:74:8a:77:f7:34:ad:4f

```


Now you can use `rekor-cli` to inspect the tlog entry

```bash
$ rekor-cli search --rekor_server https://rekor.sigstore.dev    --sha  da86794cea5a72a0ca216ca9c7ad9b38b1c50e5604a71762bcac58bab059cbe8
Found matching entries (listed by UUID):
108e9186e8c5677aba3a779854b6365ddda19dff8a869434f2f204f78596bd17f02ee5383e53f51c
108e9186e8c5677a968492cdcb7a6ecff4a5e0c945b994c4c27c2973693c8f48cf81a635e70780c2
```

for our log entry

```bash
$ rekor-cli get --rekor_server https://rekor.sigstore.dev    --uuid 108e9186e8c5677a2405e4f29eb9bfd6483073f2adb8a54f08cc8cd70d60c4e344570a5c568d55d4 

LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 947581057
IntegratedTime: 2026-02-13T15:09:35Z
UUID: 108e9186e8c5677a2405e4f29eb9bfd6483073f2adb8a54f08cc8cd70d60c4e344570a5c568d55d4
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "da86794cea5a72a0ca216ca9c7ad9b38b1c50e5604a71762bcac58bab059cbe8"
      }
    },
    "signature": {
      "content": "MEUCIFy5OPhAZa2AGGtF9I+fUXP6vs5+DUWgnmLiyCiduxOgAiEAvoFkpgnn8kTVQHHDdko+fXDhe8q559g5rmgi6Y91y+c=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMVENDQnJPZ0F3SUJBZ0lVT3ByR1puRVZySGpkMHZCTzN2VG4xZzF4ZXdVd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qRXpNVFV3T1RNMFdoY05Nall3TWpFek1UVXhPVE0wV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVSNmR1aFJtekhCSGdrRDRQbnhUOUs2VFJxQTFta2d6Nk5jNkQKQzkzc2VubU5zdkdtNFdrckQ3N05sOUorMkd5Tyt5d3BLWUZnN1laUUNIT2RBY3RNMUtPQ0JkSXdnZ1hPTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVU3bnBFCnY2U0laamxnemhxbXExeWQybWFFRUs0d0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJMU1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDaGlaR0ZqWm1Sa01tTTRZamt5TW1OaFkyRTJOalZpCk16QmlaVFUwWW1ZMU1tSmxZek16WlRjM01CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5TlRBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNalV3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2hpWkdGalptUmsKTW1NNFlqa3lNbU5oWTJFMk5qVmlNekJpWlRVMFltWTFNbUpsWXpNelpUYzNNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b1ltUmhZMlprWkRKak9HSTVNakpqWVdOaE5qWTFZak13WW1VMU5HSm1OVEppWldNegpNMlUzTnpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakkxTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlOVEE0QmdvckJnRUVBWU8vTUFFVEJDb01LR0prWVdObQpaR1F5WXpoaU9USXlZMkZqWVRZMk5XSXpNR0psTlRSaVpqVXlZbVZqTXpObE56Y3dGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl4T1RreE5qazFOelUyTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZa0dDaXNHQVFRQjFua0NCQUlFZXdSNUFIY0FkUURkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp4WGpOb2hBQUFFQXdCR01FUUNJQmJid0FoZnlCQkF1UUMyc090ZnNMZVZ3VlJadGZtRwpzYTZxWkV1eHJRMXlBaUJpZUYyeXE0Ulh5YWducmlFQTZ3OEsvNWhRRkdCQnM4bnJkM0daWHBZNVh6QUtCZ2dxCmhrak9QUVFEQXdOb0FEQmxBakJyZ1NoSmc5OExpa3Z6KytRY0JleVJZRzVIOGFDV05kTk9mOHlnd1Jyc29tTUEKZi9YdkZQdkhyS0hTdGxuSGpQRUNNUUMxYkhSbEhSYVlPVC9VVlRhN0p3YXh4ZDlWTnN0ankrd05lUHUzbitZYwpIdlp1eEFzeGtmREpSM1NLZC9jMHJVOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

you can use crane to inspect the signature too

```bash
$ crane  manifest salrashid123/server_image:sha256-bbfdc97be6fbabb8be57c346c47816c8a95904b14448270cd0585a822c33961a.sig | jq '.'


{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 233,
    "digest": "sha256:b8156ee8ea88b6b80041174a39c3ea45048598b4b7f5f15ee6e2d246671559bf"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:da86794cea5a72a0ca216ca9c7ad9b38b1c50e5604a71762bcac58bab059cbe8",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEUCIFy5OPhAZa2AGGtF9I+fUXP6vs5+DUWgnmLiyCiduxOgAiEAvoFkpgnn8kTVQHHDdko+fXDhe8q559g5rmgi6Y91y+c=",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEUCICvDQTw6TSI3GgrhqzH+AOy0A2k5mMtzoDleadAywHCNAiEAp2DIhQpZQ4Y1vWdB5eRrJZTtA8IubLOLgZM/SWtKqGA=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJkYTg2Nzk0Y2VhNWE3MmEwY2EyMTZjYTljN2FkOWIzOGIxYzUwZTU2MDRhNzE3NjJiY2FjNThiYWIwNTljYmU4In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRnk1T1BoQVphMkFHR3RGOUkrZlVYUDZ2czUrRFVXZ25tTGl5Q2lkdXhPZ0FpRUF2b0ZrcGdubjhrVFZRSEhEZGtvK2ZYRGhlOHE1NTlnNXJtZ2k2WTkxeStjPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1WRU5EUW5KUFowRjNTVUpCWjBsVlQzQnlSMXB1UlZaeVNHcGtNSFpDVHpOMlZHNHhaekY0WlhkVmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUlhwTlZGVjNUMVJOTUZkb1kwNU5hbGwzVFdwRmVrMVVWWGhQVkUwd1YycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVZTTm1SMWFGSnRla2hDU0dkclJEUlFibmhVT1VzMlZGSnhRVEZ0YTJkNk5rNWpOa1FLUXpremMyVnViVTV6ZGtkdE5GZHJja1EzTjA1c09Vb3JNa2Q1VHl0NWQzQkxXVVpuTjFsYVVVTklUMlJCWTNSTk1VdFBRMEprU1hkbloxaFBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlUzYm5CRkNuWTJVMGxhYW14bmVtaHhiWEV4ZVdReWJXRkZSVXMwZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTVUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRGFHbGFSMFpxV20xU2EwMXRUVFJaYW10NVRXMU9hRmt5UlRKT2FsWnBDazE2UW1sYVZGVXdXVzFaTVUxdFNteFplazE2V2xSak0wMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUbFJCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFsVjNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJocFdrZEdhbHB0VW1zS1RXMU5ORmxxYTNsTmJVNW9XVEpGTWs1cVZtbE5la0pwV2xSVk1GbHRXVEZOYlVwc1dYcE5lbHBVWXpOTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMWx0VW1oWk1scHJXa1JLYWs5SFNUVk5ha3BxV1ZkT2FFNXFXVEZaYWsxM1dXMVZNVTVIU20xT1ZFcHBXbGROZWdwTk1sVXpUbnBCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2t4VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxPVkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSMHByV1ZkT2JRcGFSMUY1V1hwb2FVOVVTWGxaTWtacVdWUlpNazVYU1hwTlIwcHNUbFJTYVZwcVZYbFpiVlpxVFhwT2JFNTZZM2RHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDRUMVJyZUU1cWF6Rk9lbFV5VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFphMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNFdHcE9iMmhCUVVGRlFYZENSMDFGVVVOSlFtSmlkMEZvWm5sQ1FrRjFVVU15YzA5MFpuTk1aVlozVmxKYWRHWnRSd3B6WVRaeFdrVjFlSEpSTVhsQmFVSnBaVVl5ZVhFMFVsaDVZV2R1Y21sRlFUWjNPRXN2TldoUlJrZENRbk00Ym5Ka00wZGFXSEJaTlZoNlFVdENaMmR4Q21ocmFrOVFVVkZFUVhkT2IwRkVRbXhCYWtKeVoxTm9TbWM1T0V4cGEzWjZLeXRSWTBKbGVWSlpSelZJT0dGRFYwNWtUazltT0hsbmQxSnljMjl0VFVFS1ppOVlka1pRZGtoeVMwaFRkR3h1U0dwUVJVTk5VVU14WWtoU2JFaFNZVmxQVkM5VlZsUmhOMHAzWVhoNFpEbFdUbk4wYW5rcmQwNWxVSFV6Yml0Wll3cElkbHAxZUVGemVHdG1SRXBTTTFOTFpDOWpNSEpWT0QwS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19\",\"integratedTime\":1770995375,\"logIndex\":947581057,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
        "dev.sigstore.cosign/certificate": "-----BEGIN CERTIFICATE-----\nMIIHLTCCBrOgAwIBAgIUOprGZnEVrHjd0vBO3vTn1g1xewUwCgYIKoZIzj0EAwMw\nNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl\ncm1lZGlhdGUwHhcNMjYwMjEzMTUwOTM0WhcNMjYwMjEzMTUxOTM0WjAAMFkwEwYH\nKoZIzj0CAQYIKoZIzj0DAQcDQgAER6duhRmzHBHgkD4PnxT9K6TRqA1mkgz6Nc6D\nC93senmNsvGm4WkrD77Nl9J+2GyO+ywpKYFg7YZQCHOdActM1KOCBdIwggXOMA4G\nA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQU7npE\nv6SIZjlgzhqmq1yd2maEEK4wHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y\nZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy\nMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs\nZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI1MDkGCisGAQQBg78wAQEEK2h0dHBz\nOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD\nvzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBChiZGFjZmRkMmM4YjkyMmNhY2E2NjVi\nMzBiZTU0YmY1MmJlYzMzZTc3MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB\nBAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf\nBgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yNTA7BgorBgEEAYO/MAEIBC0M\nK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK\nKwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv\nLWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl\nLnlhbWxAcmVmcy90YWdzL3YwLjAuMjUwOAYKKwYBBAGDvzABCgQqDChiZGFjZmRk\nMmM4YjkyMmNhY2E2NjViMzBiZTU0YmY1MmJlYzMzZTc3MB0GCisGAQQBg78wAQsE\nDwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi\nLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG\nAQQBg78wAQ0EKgwoYmRhY2ZkZDJjOGI5MjJjYWNhNjY1YjMwYmU1NGJmNTJiZWMz\nM2U3NzAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI1MBoGCisGAQQB\ng78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0\naHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5\nBgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv\nZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh\nc2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yNTA4BgorBgEEAYO/MAETBCoMKGJkYWNm\nZGQyYzhiOTIyY2FjYTY2NWIzMGJlNTRiZjUyYmVjMzNlNzcwFAYKKwYBBAGDvzAB\nFAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh\nbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z\nLzIxOTkxNjk1NzU2L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw\ngYkGCisGAQQB1nkCBAIEewR5AHcAdQDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H\ninKALynujgAAAZxXjNohAAAEAwBGMEQCIBbbwAhfyBBAuQC2sOtfsLeVwVRZtfmG\nsa6qZEuxrQ1yAiBieF2yq4RXyagnriEA6w8K/5hQFGBBs8nrd3GZXpY5XzAKBggq\nhkjOPQQDAwNoADBlAjBrgShJg98Likvz++QcBeyRYG5H8aCWNdNOf8ygwRrsomMA\nf/XvFPvHrKHStlnHjPECMQC1bHRlHRaYOT/UVTa7Jwaxxd9VNstjy+wNePu3n+Yc\nHvZuxAsxkfDJR3SKd/c0rU8=\n-----END CERTIFICATE-----\n",
        "dev.sigstore.cosign/chain": "-----BEGIN CERTIFICATE-----\nMIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRkHrn/UnfaQwCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMjA0MTMyMDA2MTVaFw0zMTEwMDUxMzU2NThaMDcxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjEeMBwGA1UEAxMVc2lnc3RvcmUtaW50ZXJtZWRpYXRlMHYwEAYHKoZIzj0C\nAQYFK4EEACIDYgAE8RVS/ysH+NOvuDZyPIZtilgUF9NlarYpAd9HP1vBBH1U5CV7\n7LSS7s0ZiH4nE7Hv7ptS6LvvR/STk798LVgMzLlJ4HeIfF3tHSaexLcYpSASr1kS\n0N/RgBJz/9jWCiXno3sweTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYB\nBQUHAwMwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU39Ppz1YkEZb5qNjp\nKFWixi4YZD8wHwYDVR0jBBgwFoAUWMAeX5FFpWapesyQoZMi0CrFxfowCgYIKoZI\nzj0EAwMDZwAwZAIwPCsQK4DYiZYDPIaDi5HFKnfxXx6ASSVmERfsynYBiX2X6SJR\nnZU84/9DZdnFvvxmAjBOt6QpBlc4J/0DxvkTCqpclvziL6BCCPnjdlIB3Pu3BxsP\nmygUY7Ii2zbdCdliiow=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIB9zCCAXygAwIBAgIUALZNAPFdxHPwjeDloDwyYChAO/4wCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMTEwMDcxMzU2NTlaFw0zMTEwMDUxMzU2NThaMCoxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjERMA8GA1UEAxMIc2lnc3RvcmUwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT7\nXeFT4rb3PQGwS4IajtLk3/OlnpgangaBclYpsYBr5i+4ynB07ceb3LP0OIOZdxex\nX69c5iVuyJRQ+Hz05yi+UF3uBWAlHpiS5sh0+H2GHE7SXrk1EC5m1Tr19L9gg92j\nYzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRY\nwB5fkUWlZql6zJChkyLQKsXF+jAfBgNVHSMEGDAWgBRYwB5fkUWlZql6zJChkyLQ\nKsXF+jAKBggqhkjOPQQDAwNpADBmAjEAj1nHeXZp+13NWBNa+EDsDP8G1WWg1tCM\nWP/WHPqpaVo0jhsweNFZgSs0eE7wYI4qAjEA2WB9ot98sIkoF3vZYdd3/VtWB5b9\nTNMea7Ix/stJ5TfcLLeABLE4BNJOsQ4vnBHJ\n-----END CERTIFICATE-----"
      }
    }
  ]
}
```


#### Verify binary

TODO: make this work with github workflows...right now it just uses key pairs

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
      "digest": "AMUJvfQjzaU/jlIuvIywXTOqgos4Gena9/BdXziaJnE="
    },
    "signature": "hbMwEXSVAyqid57566N4LSI5NF6PwewSAjd6maVSPTU3GskRzMkctTCeu3TieIAW2GDB4Z0GoeABxBDybuoGIQIowMSh7x+ufqB62O7eVB38e29rrt2CFEW/zXonSCPga6bzkXE0aKteiN+iFB7dtYOnsmPjNcZMYuycLvia1QtpYuzvHgC9se/c+WeP4P1BT8RGnaByLI6B+URwU0+lPAa0cRaDUWF3E8RZKBU3+vhk6W9+/PCaknDCP6g4X/eI1jNB5AJw1B/CsvDXIt7aW8dYaSPP1bgsFwKtIsCAoxYr4bMkDa695AVI9+Ud5rbcCaR3DO4ZM+iNffJ41CRs4w=="
  }
}


### which can be verified locally with the local signing key
export sig=`cat bazel-bin/app/server_linux_amd64.sig | jq -r '.messageSignature.signature'`
cosign verify-blob --key certs/import-cosign.pub --signature $sig bazel-bin/app/server_linux_amd64_bin
```