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

export IMAGE="docker.io/salrashid123/server_image:server@sha256:d454b76c23edb9f9abf0541257dc0e92ee9c16df25d4676ec0946f6beae12ef5"

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
export IMAGE="docker.io/salrashid123/server_image:server@sha256:adce22cdd04fa2b012209f2d048c883c66b7ab89eaadf4f596899fde52d3b0bd"

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

export TAG=v0.0.19
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

Verification for index.docker.io/salrashid123/server_image@sha256:adce22cdd04fa2b012209f2d048c883c66b7ab89eaadf4f596899fde52d3b0bd --
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
        "docker-manifest-digest": "sha256:adce22cdd04fa2b012209f2d048c883c66b7ab89eaadf4f596899fde52d3b0bd"
      },
      "type": "cosign container image signature"
    },
    "optional": {
      "1.3.6.1.4.1.57264.1.1": "https://token.actions.githubusercontent.com",
      "1.3.6.1.4.1.57264.1.2": "push",
      "1.3.6.1.4.1.57264.1.3": "291dab9af96caa83aef0968dc1f757039d3c166b",
      "1.3.6.1.4.1.57264.1.4": "Release",
      "1.3.6.1.4.1.57264.1.5": "salrashid123/go-bazel-github-workflow",
      "1.3.6.1.4.1.57264.1.6": "refs/tags/v0.0.14",
      "Bundle": {
        "SignedEntryTimestamp": "MEUCIDb5ngT0cYnECAPpZR88M57VSt2SMuCw32+6cWQDy0m7AiEAz5q89Wb2LzWeun3luNuOsKCxWdQzwQvmKnRVR/8JS14=",
        "Payload": {
          "body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJhZjliMjk5MmFkZmQxMzRkZjE0NjdmMWY5YWI0YWI1MTY3OThmYTljZWNmMmU4NTdjYTA2YzEzOWZiOThiMzRmIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRnArU3YvNWlpZUEwN0FLRTF0MjltdzFUSFZFckg4SUxDbHdvS1lhTWo4c0FpRUF4R3FGTHBmcUJmVmR4N1pYNElRaUg1eDd6emhDVXZWL1VvUDUxNXVVWUVNPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1ha05EUW5KVFowRjNTVUpCWjBsVlpFOVRjbUUzYmpWQ2FGTlJhVWhDYWxSaFRtUXhORTVMTlZwSmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUVRKTlZGRXdUVlJGTVZkb1kwNU5hbGwzVFdwQk1rMVVVVEZOVkVVeFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVYzVDNOYU9XaEtaVzg1V1hCeFl6WnJabkU1V2l0SGEzUTNWMFl3TkVSeGIyZDRhemdLWmxWRmIyRmpTblJ6WmpGcE4ySTJUVzlKTkdWbVlYQjZiVGRSY25CT1pGcG1LMFJXUTA1dUsxUkdkMVZaZUhOWFNrdFBRMEprVFhkbloxaFFUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlZuTjNaUUNsaGxja2hsZHk5V1NUUkhhWEprYkhkYVVuZEthRmh2ZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BGTUUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRFozbFBWRVpyV1ZkSk5WbFhXVFZPYlU1b1dWUm5lbGxYVm0xTlJHc3lDazlIVW1wTlYxa3pUbFJqZDAxNmJHdE5NazE0VG1wYWFVMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjRUa1JCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTlZGRjNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJkNVQxUkdhMWxYU1RVS1dWZFpOVTV0VG1oWlZHZDZXVmRXYlUxRWF6SlBSMUpxVFZkWk0wNVVZM2ROZW14clRUSk5lRTVxV21sTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMDFxYTNoYVIwWnBUMWRHYlU5VVdtcFpWMFUwVFRKR2JGcHFRVFZPYW1ocldYcEdiVTU2VlROTlJFMDFXa1JPYWdwTlZGa3lXV3BCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha1V3VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGhPUkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSRWsxVFZkU2FBcFphbXhvV21wck1sa3lSbWhQUkU1b1dsZFpkMDlVV1RSYVIwMTRXbXBqTVU1NlFYcFBWMUY2V1hwRk1rNXRTWGRHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDRUbnBWTUUxNlFUUlBWR015VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFpiMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaa0ZTTmtGSVowRmtaMFJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwM2VscHRaR1JCUVVGRlFYZENTRTFGVlVOSlIxb3hhbXRYS3pKeVlsQXZSRzFJSzBNdk1GZG9SbFZLYVZOT1ZHY3ZWUXBoYkhZNWIzWkJNQ3RFVTNkQmFVVkJiRGw0VUU5YVpUUktkMjVwY0ZkcFJrRkJPVXRUYUV0bVpWRjRMMmhhWjFKbFJuTlFaWFJHWmxkeFZYZERaMWxKQ2t0dldrbDZhakJGUVhkTlJHRkJRWGRhVVVsNFFVMHlOVmgyVDBwSlVqQXdWVlYzT1hwc00yZFZaVVpDUjBOa0t6UlFaREo1Tm5oSGFIaFJhMk16UkVrS2NtWktTMVl5ZHl0dFptVkZLMVJ0SzFJd2FWTktaMGwzVWl0NlZuVnJVWFYxTW0xc2FYRlhVWFZaY1ZKYU5WVm1jMVZET1RoTmNERmxNMGhaT1hob1RRcFVVR2RKZEhGa01qZHRSbUk1ZW5JeFFtTk1VV2hXZERVS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19",
          "integratedTime": 1770388876,
          "logIndex": 924321775,
          "logID": "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
        }
      },
      "Issuer": "https://token.actions.githubusercontent.com",
      "Subject": "https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.14",
      "githubWorkflowName": "Release",
      "githubWorkflowRef": "refs/tags/v0.0.14",
      "githubWorkflowRepository": "salrashid123/go-bazel-github-workflow",
      "githubWorkflowSha": "291dab9af96caa83aef0968dc1f757039d3c166b",
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
        "value": "af9b2992adfd134df1467f1f9ab4ab516798fa9cecf2e857ca06c139fb98b34f"
      }
    },
    "signature": {
      "content": "MEUCIFp+Sv/5iieA07AKE1t29mw1THVErH8ILClwoKYaMj8sAiEAxGqFLpfqBfVdx7ZX4IQiH5x7zzhCUvV/UoP515uUYEM=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMakNDQnJTZ0F3SUJBZ0lVZE9TcmE3bjVCaFNRaUhCalRhTmQxNE5LNVpJd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qQTJNVFEwTVRFMVdoY05Nall3TWpBMk1UUTFNVEUxV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUV3T3NaOWhKZW85WXBxYzZrZnE5WitHa3Q3V0YwNERxb2d4azgKZlVFb2FjSnRzZjFpN2I2TW9JNGVmYXB6bTdRcnBOZFpmK0RWQ05uK1RGd1VZeHNXSktPQ0JkTXdnZ1hQTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVVnN3ZQClhlckhldy9WSTRHaXJkbHdaUndKaFhvd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpFME1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDZ3lPVEZrWVdJNVlXWTVObU5oWVRnellXVm1NRGsyCk9HUmpNV1kzTlRjd016bGtNMk14TmpaaU1CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR4TkRBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNVFF3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2d5T1RGa1lXSTUKWVdZNU5tTmhZVGd6WVdWbU1EazJPR1JqTVdZM05UY3dNemxrTTJNeE5qWmlNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b01qa3haR0ZpT1dGbU9UWmpZV0U0TTJGbFpqQTVOamhrWXpGbU56VTNNRE01WkROagpNVFkyWWpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakUwTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHhOREE0QmdvckJnRUVBWU8vTUFFVEJDb01LREk1TVdSaApZamxoWmprMlkyRmhPRE5oWldZd09UWTRaR014WmpjMU56QXpPV1F6WXpFMk5tSXdGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl4TnpVME16QTRPVGMyTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZb0dDaXNHQVFRQjFua0NCQUlFZkFSNkFIZ0FkZ0RkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp3elptZGRBQUFFQXdCSE1FVUNJR1oxamtXKzJyYlAvRG1IK0MvMFdoRlVKaVNOVGcvVQphbHY5b3ZBMCtEU3dBaUVBbDl4UE9aZTRKd25pcFdpRkFBOUtTaEtmZVF4L2haZ1JlRnNQZXRGZldxVXdDZ1lJCktvWkl6ajBFQXdNRGFBQXdaUUl4QU0yNVh2T0pJUjAwVVV3OXpsM2dVZUZCR0NkKzRQZDJ5NnhHaHhRa2MzREkKcmZKS1YydyttZmVFK1RtK1IwaVNKZ0l3Uit6VnVrUXV1Mm1saXFXUXVZcVJaNVVmc1VDOThNcDFlM0hZOXhoTQpUUGdJdHFkMjdtRmI5enIxQmNMUWhWdDUKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

the content is the public key

```bash
cat public.crt

-----BEGIN CERTIFICATE-----
MIIHLjCCBrSgAwIBAgIUdOSra7n5BhSQiHBjTaNd14NK5ZIwCgYIKoZIzj0EAwMw
NzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl
cm1lZGlhdGUwHhcNMjYwMjA2MTQ0MTE1WhcNMjYwMjA2MTQ1MTE1WjAAMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAEwOsZ9hJeo9Ypqc6kfq9Z+Gkt7WF04Dqogxk8
fUEoacJtsf1i7b6MoI4efapzm7QrpNdZf+DVCNn+TFwUYxsWJKOCBdMwggXPMA4G
A1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQUg7vP
XerHew/VI4GirdlwZRwJhXowHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y
ZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy
My9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs
ZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjE0MDkGCisGAQQBg78wAQEEK2h0dHBz
Oi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD
vzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBCgyOTFkYWI5YWY5NmNhYTgzYWVmMDk2
OGRjMWY3NTcwMzlkM2MxNjZiMBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB
BAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf
BgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4xNDA7BgorBgEEAYO/MAEIBC0M
K2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK
KwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv
LWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl
LnlhbWxAcmVmcy90YWdzL3YwLjAuMTQwOAYKKwYBBAGDvzABCgQqDCgyOTFkYWI5
YWY5NmNhYTgzYWVmMDk2OGRjMWY3NTcwMzlkM2MxNjZiMB0GCisGAQQBg78wAQsE
DwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi
LmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG
AQQBg78wAQ0EKgwoMjkxZGFiOWFmOTZjYWE4M2FlZjA5NjhkYzFmNzU3MDM5ZDNj
MTY2YjAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjE0MBoGCisGAQQB
g78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0
aHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5
BgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv
Z28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh
c2UueWFtbEByZWZzL3RhZ3MvdjAuMC4xNDA4BgorBgEEAYO/MAETBCoMKDI5MWRh
YjlhZjk2Y2FhODNhZWYwOTY4ZGMxZjc1NzAzOWQzYzE2NmIwFAYKKwYBBAGDvzAB
FAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh
bHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z
LzIxNzU0MzA4OTc2L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw
gYoGCisGAQQB1nkCBAIEfAR6AHgAdgDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H
inKALynujgAAAZwzZmddAAAEAwBHMEUCIGZ1jkW+2rbP/DmH+C/0WhFUJiSNTg/U
alv9ovA0+DSwAiEAl9xPOZe4JwnipWiFAA9KShKfeQx/hZgReFsPetFfWqUwCgYI
KoZIzj0EAwMDaAAwZQIxAM25XvOJIR00UUw9zl3gUeFBGCd+4Pd2y6xGhxQkc3DI
rfJKV2w+mfeE+Tm+R0iSJgIwR+zVukQuu2mliqWQuYqRZ5UfsUC98Mp1e3HY9xhM
TPgItqd27mFb9zr1BcLQhVt5
-----END CERTIFICATE-----
```


which includes

```bash
$  openssl x509 -in public.crt -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            74:e4:ab:6b:b9:f9:06:14:90:88:70:63:4d:a3:5d:d7:83:4a:e5:92
        Signature Algorithm: ecdsa-with-SHA384
        Issuer: O=sigstore.dev, CN=sigstore-intermediate
        Validity
            Not Before: Feb  6 14:41:15 2026 GMT
            Not After : Feb  6 14:51:15 2026 GMT
        Subject: 
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:c0:eb:19:f6:12:5e:a3:d6:29:a9:ce:a4:7e:af:
                    59:f8:69:2d:ed:61:74:e0:3a:a8:83:19:3c:7d:41:
                    28:69:c2:6d:b1:fd:62:ed:be:8c:a0:8e:1e:7d:aa:
                    73:9b:b4:2b:a4:d7:59:7f:e0:d5:08:d9:fe:4c:5c:
                    14:63:1b:16:24
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage: 
                Code Signing
            X509v3 Subject Key Identifier: 
                83:BB:CF:5D:EA:C7:7B:0F:D5:23:81:A2:AD:D9:70:65:1C:09:85:7A
            X509v3 Authority Key Identifier: 
                DF:D3:E9:CF:56:24:11:96:F9:A8:D8:E9:28:55:A2:C6:2E:18:64:3F
            X509v3 Subject Alternative Name: critical
                URI:https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.14
            1.3.6.1.4.1.57264.1.1: 
                https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.2: 
                push
            1.3.6.1.4.1.57264.1.3: 
                291dab9af96caa83aef0968dc1f757039d3c166b
            1.3.6.1.4.1.57264.1.4: 
                Release
            1.3.6.1.4.1.57264.1.5: 
                salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.6: 
                refs/tags/v0.0.14
            1.3.6.1.4.1.57264.1.8: 
                .+https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.9: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.14
            1.3.6.1.4.1.57264.1.10: 
                .(291dab9af96caa83aef0968dc1f757039d3c166b
            1.3.6.1.4.1.57264.1.11: 
github-hosted   .
            1.3.6.1.4.1.57264.1.12: 
                .8https://github.com/salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.13: 
                .(291dab9af96caa83aef0968dc1f757039d3c166b
            1.3.6.1.4.1.57264.1.14: 
                ..refs/tags/v0.0.14
            1.3.6.1.4.1.57264.1.15: 
                .
1150755799
            1.3.6.1.4.1.57264.1.16: 
                ..https://github.com/salrashid123
            1.3.6.1.4.1.57264.1.17: 
                ..11149054
            1.3.6.1.4.1.57264.1.18: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.14
            1.3.6.1.4.1.57264.1.19: 
                .(291dab9af96caa83aef0968dc1f757039d3c166b
            1.3.6.1.4.1.57264.1.20: 
                ..push
            1.3.6.1.4.1.57264.1.21: 
                .\https://github.com/salrashid123/go-bazel-github-workflow/actions/runs/21754308976/attempts/1
            1.3.6.1.4.1.57264.1.22: 
                ..public
            CT Precertificate SCTs: 
                Signed Certificate Timestamp:
                    Version   : v1 (0x0)
                    Log ID    : DD:3D:30:6A:C6:C7:11:32:63:19:1E:1C:99:67:37:02:
                                A2:4A:5E:B8:DE:3C:AD:FF:87:8A:72:80:2F:29:EE:8E
                    Timestamp : Feb  6 14:41:15.101 2026 GMT
                    Extensions: none
                    Signature : ecdsa-with-SHA256
                                30:45:02:20:66:75:8E:45:BE:DA:B6:CF:FC:39:87:F8:
                                2F:F4:5A:11:54:26:24:8D:4E:0F:D4:6A:5B:FD:A2:F0:
                                34:F8:34:B0:02:21:00:97:DC:4F:39:97:B8:27:09:E2:
                                A5:68:85:00:0F:4A:4A:12:9F:79:0C:7F:85:98:11:78:
                                5B:0F:7A:D1:5F:5A:A5
    Signature Algorithm: ecdsa-with-SHA384
    Signature Value:
        30:65:02:31:00:cd:b9:5e:f3:89:21:1d:34:51:4c:3d:ce:5d:
        e0:51:e1:41:18:27:7e:e0:f7:76:cb:ac:46:87:14:24:73:70:
        c8:ad:f2:4a:57:6c:3e:99:f7:84:f9:39:be:47:48:92:26:02:
        30:47:ec:d5:ba:44:2e:bb:69:a5:8a:a5:90:b9:8a:91:67:95:
        1f:b1:40:bd:f0:ca:75:7b:71:d8:f7:18:4c:4c:f8:08:b6:a7:
        76:ee:61:5b:f7:3a:f5:05:c2:d0:85:5b:79
```


Now you can use `rekor-cli` to inspect the tlog entry

```bash
$ rekor-cli search --rekor_server https://rekor.sigstore.dev    --sha  af9b2992adfd134df1467f1f9ab4ab516798fa9cecf2e857ca06c139fb98b34f
Found matching entries (listed by UUID):
108e9186e8c5677aba3a779854b6365ddda19dff8a869434f2f204f78596bd17f02ee5383e53f51c
108e9186e8c5677a968492cdcb7a6ecff4a5e0c945b994c4c27c2973693c8f48cf81a635e70780c2
```

for our log entry

```bash
$ rekor-cli get --rekor_server https://rekor.sigstore.dev    --uuid 108e9186e8c5677a968492cdcb7a6ecff4a5e0c945b994c4c27c2973693c8f48cf81a635e70780c2 

LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 924321775
IntegratedTime: 2026-02-06T14:41:16Z
UUID: 108e9186e8c5677a968492cdcb7a6ecff4a5e0c945b994c4c27c2973693c8f48cf81a635e70780c2
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "af9b2992adfd134df1467f1f9ab4ab516798fa9cecf2e857ca06c139fb98b34f"
      }
    },
    "signature": {
      "content": "MEUCIFp+Sv/5iieA07AKE1t29mw1THVErH8ILClwoKYaMj8sAiEAxGqFLpfqBfVdx7ZX4IQiH5x7zzhCUvV/UoP515uUYEM=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMakNDQnJTZ0F3SUJBZ0lVZE9TcmE3bjVCaFNRaUhCalRhTmQxNE5LNVpJd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qQTJNVFEwTVRFMVdoY05Nall3TWpBMk1UUTFNVEUxV2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUV3T3NaOWhKZW85WXBxYzZrZnE5WitHa3Q3V0YwNERxb2d4azgKZlVFb2FjSnRzZjFpN2I2TW9JNGVmYXB6bTdRcnBOZFpmK0RWQ05uK1RGd1VZeHNXSktPQ0JkTXdnZ1hQTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVVnN3ZQClhlckhldy9WSTRHaXJkbHdaUndKaFhvd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpFME1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDZ3lPVEZrWVdJNVlXWTVObU5oWVRnellXVm1NRGsyCk9HUmpNV1kzTlRjd016bGtNMk14TmpaaU1CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR4TkRBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNVFF3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2d5T1RGa1lXSTUKWVdZNU5tTmhZVGd6WVdWbU1EazJPR1JqTVdZM05UY3dNemxrTTJNeE5qWmlNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b01qa3haR0ZpT1dGbU9UWmpZV0U0TTJGbFpqQTVOamhrWXpGbU56VTNNRE01WkROagpNVFkyWWpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakUwTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHhOREE0QmdvckJnRUVBWU8vTUFFVEJDb01LREk1TVdSaApZamxoWmprMlkyRmhPRE5oWldZd09UWTRaR014WmpjMU56QXpPV1F6WXpFMk5tSXdGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl4TnpVME16QTRPVGMyTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZb0dDaXNHQVFRQjFua0NCQUlFZkFSNkFIZ0FkZ0RkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp3elptZGRBQUFFQXdCSE1FVUNJR1oxamtXKzJyYlAvRG1IK0MvMFdoRlVKaVNOVGcvVQphbHY5b3ZBMCtEU3dBaUVBbDl4UE9aZTRKd25pcFdpRkFBOUtTaEtmZVF4L2haZ1JlRnNQZXRGZldxVXdDZ1lJCktvWkl6ajBFQXdNRGFBQXdaUUl4QU0yNVh2T0pJUjAwVVV3OXpsM2dVZUZCR0NkKzRQZDJ5NnhHaHhRa2MzREkKcmZKS1YydyttZmVFK1RtK1IwaVNKZ0l3Uit6VnVrUXV1Mm1saXFXUXVZcVJaNVVmc1VDOThNcDFlM0hZOXhoTQpUUGdJdHFkMjdtRmI5enIxQmNMUWhWdDUKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

you can use crane to inspect the signature too

```bash
$ crane  manifest salrashid123/server_image:sha256-adce22cdd04fa2b012209f2d048c883c66b7ab89eaadf4f596899fde52d3b0bd.sig | jq '.'


{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 233,
    "digest": "sha256:a34c0b7df5831bf35493245a6389c9eee562e976e21e1587684fa92c6eb8c094"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:af9b2992adfd134df1467f1f9ab4ab516798fa9cecf2e857ca06c139fb98b34f",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEUCIFp+Sv/5iieA07AKE1t29mw1THVErH8ILClwoKYaMj8sAiEAxGqFLpfqBfVdx7ZX4IQiH5x7zzhCUvV/UoP515uUYEM=",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEUCIDb5ngT0cYnECAPpZR88M57VSt2SMuCw32+6cWQDy0m7AiEAz5q89Wb2LzWeun3luNuOsKCxWdQzwQvmKnRVR/8JS14=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJhZjliMjk5MmFkZmQxMzRkZjE0NjdmMWY5YWI0YWI1MTY3OThmYTljZWNmMmU4NTdjYTA2YzEzOWZiOThiMzRmIn19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRnArU3YvNWlpZUEwN0FLRTF0MjltdzFUSFZFckg4SUxDbHdvS1lhTWo4c0FpRUF4R3FGTHBmcUJmVmR4N1pYNElRaUg1eDd6emhDVXZWL1VvUDUxNXVVWUVNPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1ha05EUW5KVFowRjNTVUpCWjBsVlpFOVRjbUUzYmpWQ2FGTlJhVWhDYWxSaFRtUXhORTVMTlZwSmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUVRKTlZGRXdUVlJGTVZkb1kwNU5hbGwzVFdwQk1rMVVVVEZOVkVVeFYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVYzVDNOYU9XaEtaVzg1V1hCeFl6WnJabkU1V2l0SGEzUTNWMFl3TkVSeGIyZDRhemdLWmxWRmIyRmpTblJ6WmpGcE4ySTJUVzlKTkdWbVlYQjZiVGRSY25CT1pGcG1LMFJXUTA1dUsxUkdkMVZaZUhOWFNrdFBRMEprVFhkbloxaFFUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlZuTjNaUUNsaGxja2hsZHk5V1NUUkhhWEprYkhkYVVuZEthRmh2ZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BGTUUxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRFozbFBWRVpyV1ZkSk5WbFhXVFZPYlU1b1dWUm5lbGxYVm0xTlJHc3lDazlIVW1wTlYxa3pUbFJqZDAxNmJHdE5NazE0VG1wYWFVMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjRUa1JCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTlZGRjNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJkNVQxUkdhMWxYU1RVS1dWZFpOVTV0VG1oWlZHZDZXVmRXYlUxRWF6SlBSMUpxVFZkWk0wNVVZM2ROZW14clRUSk5lRTVxV21sTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMDFxYTNoYVIwWnBUMWRHYlU5VVdtcFpWMFUwVFRKR2JGcHFRVFZPYW1ocldYcEdiVTU2VlROTlJFMDFXa1JPYWdwTlZGa3lXV3BCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha1V3VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGhPUkVFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSRWsxVFZkU2FBcFphbXhvV21wck1sa3lSbWhQUkU1b1dsZFpkMDlVV1RSYVIwMTRXbXBqTVU1NlFYcFBWMUY2V1hwRk1rNXRTWGRHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDRUbnBWTUUxNlFUUlBWR015VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFpiMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaa0ZTTmtGSVowRmtaMFJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwM2VscHRaR1JCUVVGRlFYZENTRTFGVlVOSlIxb3hhbXRYS3pKeVlsQXZSRzFJSzBNdk1GZG9SbFZLYVZOT1ZHY3ZWUXBoYkhZNWIzWkJNQ3RFVTNkQmFVVkJiRGw0VUU5YVpUUktkMjVwY0ZkcFJrRkJPVXRUYUV0bVpWRjRMMmhhWjFKbFJuTlFaWFJHWmxkeFZYZERaMWxKQ2t0dldrbDZhakJGUVhkTlJHRkJRWGRhVVVsNFFVMHlOVmgyVDBwSlVqQXdWVlYzT1hwc00yZFZaVVpDUjBOa0t6UlFaREo1Tm5oSGFIaFJhMk16UkVrS2NtWktTMVl5ZHl0dFptVkZLMVJ0SzFJd2FWTktaMGwzVWl0NlZuVnJVWFYxTW0xc2FYRlhVWFZaY1ZKYU5WVm1jMVZET1RoTmNERmxNMGhaT1hob1RRcFVVR2RKZEhGa01qZHRSbUk1ZW5JeFFtTk1VV2hXZERVS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19\",\"integratedTime\":1770388876,\"logIndex\":924321775,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
        "dev.sigstore.cosign/certificate": "-----BEGIN CERTIFICATE-----\nMIIHLjCCBrSgAwIBAgIUdOSra7n5BhSQiHBjTaNd14NK5ZIwCgYIKoZIzj0EAwMw\nNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl\ncm1lZGlhdGUwHhcNMjYwMjA2MTQ0MTE1WhcNMjYwMjA2MTQ1MTE1WjAAMFkwEwYH\nKoZIzj0CAQYIKoZIzj0DAQcDQgAEwOsZ9hJeo9Ypqc6kfq9Z+Gkt7WF04Dqogxk8\nfUEoacJtsf1i7b6MoI4efapzm7QrpNdZf+DVCNn+TFwUYxsWJKOCBdMwggXPMA4G\nA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQUg7vP\nXerHew/VI4GirdlwZRwJhXowHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y\nZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy\nMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs\nZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjE0MDkGCisGAQQBg78wAQEEK2h0dHBz\nOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD\nvzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBCgyOTFkYWI5YWY5NmNhYTgzYWVmMDk2\nOGRjMWY3NTcwMzlkM2MxNjZiMBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB\nBAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf\nBgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4xNDA7BgorBgEEAYO/MAEIBC0M\nK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK\nKwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv\nLWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl\nLnlhbWxAcmVmcy90YWdzL3YwLjAuMTQwOAYKKwYBBAGDvzABCgQqDCgyOTFkYWI5\nYWY5NmNhYTgzYWVmMDk2OGRjMWY3NTcwMzlkM2MxNjZiMB0GCisGAQQBg78wAQsE\nDwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi\nLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG\nAQQBg78wAQ0EKgwoMjkxZGFiOWFmOTZjYWE4M2FlZjA5NjhkYzFmNzU3MDM5ZDNj\nMTY2YjAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjE0MBoGCisGAQQB\ng78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0\naHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5\nBgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv\nZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh\nc2UueWFtbEByZWZzL3RhZ3MvdjAuMC4xNDA4BgorBgEEAYO/MAETBCoMKDI5MWRh\nYjlhZjk2Y2FhODNhZWYwOTY4ZGMxZjc1NzAzOWQzYzE2NmIwFAYKKwYBBAGDvzAB\nFAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh\nbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z\nLzIxNzU0MzA4OTc2L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw\ngYoGCisGAQQB1nkCBAIEfAR6AHgAdgDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H\ninKALynujgAAAZwzZmddAAAEAwBHMEUCIGZ1jkW+2rbP/DmH+C/0WhFUJiSNTg/U\nalv9ovA0+DSwAiEAl9xPOZe4JwnipWiFAA9KShKfeQx/hZgReFsPetFfWqUwCgYI\nKoZIzj0EAwMDaAAwZQIxAM25XvOJIR00UUw9zl3gUeFBGCd+4Pd2y6xGhxQkc3DI\nrfJKV2w+mfeE+Tm+R0iSJgIwR+zVukQuu2mliqWQuYqRZ5UfsUC98Mp1e3HY9xhM\nTPgItqd27mFb9zr1BcLQhVt5\n-----END CERTIFICATE-----\n",
        "dev.sigstore.cosign/chain": "-----BEGIN CERTIFICATE-----\nMIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRkHrn/UnfaQwCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMjA0MTMyMDA2MTVaFw0zMTEwMDUxMzU2NThaMDcxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjEeMBwGA1UEAxMVc2lnc3RvcmUtaW50ZXJtZWRpYXRlMHYwEAYHKoZIzj0C\nAQYFK4EEACIDYgAE8RVS/ysH+NOvuDZyPIZtilgUF9NlarYpAd9HP1vBBH1U5CV7\n7LSS7s0ZiH4nE7Hv7ptS6LvvR/STk798LVgMzLlJ4HeIfF3tHSaexLcYpSASr1kS\n0N/RgBJz/9jWCiXno3sweTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYB\nBQUHAwMwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU39Ppz1YkEZb5qNjp\nKFWixi4YZD8wHwYDVR0jBBgwFoAUWMAeX5FFpWapesyQoZMi0CrFxfowCgYIKoZI\nzj0EAwMDZwAwZAIwPCsQK4DYiZYDPIaDi5HFKnfxXx6ASSVmERfsynYBiX2X6SJR\nnZU84/9DZdnFvvxmAjBOt6QpBlc4J/0DxvkTCqpclvziL6BCCPnjdlIB3Pu3BxsP\nmygUY7Ii2zbdCdliiow=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIB9zCCAXygAwIBAgIUALZNAPFdxHPwjeDloDwyYChAO/4wCgYIKoZIzj0EAwMw\nKjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y\nMTEwMDcxMzU2NTlaFw0zMTEwMDUxMzU2NThaMCoxFTATBgNVBAoTDHNpZ3N0b3Jl\nLmRldjERMA8GA1UEAxMIc2lnc3RvcmUwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT7\nXeFT4rb3PQGwS4IajtLk3/OlnpgangaBclYpsYBr5i+4ynB07ceb3LP0OIOZdxex\nX69c5iVuyJRQ+Hz05yi+UF3uBWAlHpiS5sh0+H2GHE7SXrk1EC5m1Tr19L9gg92j\nYzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRY\nwB5fkUWlZql6zJChkyLQKsXF+jAfBgNVHSMEGDAWgBRYwB5fkUWlZql6zJChkyLQ\nKsXF+jAKBggqhkjOPQQDAwNpADBmAjEAj1nHeXZp+13NWBNa+EDsDP8G1WWg1tCM\nWP/WHPqpaVo0jhsweNFZgSs0eE7wYI4qAjEA2WB9ot98sIkoF3vZYdd3/VtWB5b9\nTNMea7Ix/stJ5TfcLLeABLE4BNJOsQ4vnBHJ\n-----END CERTIFICATE-----"
      }
    }
  ]
}
```


#### Verify binary

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
      "digest": "osUJ/y2M8qoeYTCYLlBZyHS+2UrlOyf8liuihxQLYok="
    },
    "signature": "B1XtUbHJHh/se5Np3iXMSWLrtbn29u4Q8IbXy3Y5cTXLjIKQwZO3MnSWjWOczVPOTAWesuEzdW64UQZRc1EAwYH4ZLYg0dmnXVnzEnC88T4+Ov4TmzM2RRSakxhRScQm4cGDv18GdtOKCgSzQhjruhG/ed7mqASaHfNnnqK6Yav/1eSr9c5lGjI4JMCA2LkyKbvN0Rc7/xzOlzWwRBsm0+zNol+OWVaipuSIplX4SV5WiJ7mtKe7MqarQ9ks3Fam8PBo8gcJaK3V+5ovKKYqRqTjv71+R3u2jDDMpFcm/WqvfsFDOCQlyojHVtFtjRfkJlTelGGFBqU37q/iwB4VQQ=="
  }
}

### which can be verified locally with the local signing key
export sig=`cat bazel-bin/app/server_linux_amd64.sig | jq -r '.messageSignature.signature'`
cosign verify-blob --key certs/import-cosign.pub --signature $sig bazel-bin/app/server_linux_amd64_bin
```