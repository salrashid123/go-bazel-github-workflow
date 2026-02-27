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

export IMAGE="docker.io/salrashid123/server_image:server@sha256:120ef883bd7612aee0a100297def41014bf3e1aa321a0eaf4d0ebd9855d39a11"

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
export IMAGE="docker.io/salrashid123/server_image:server@sha256:120ef883bd7612aee0a100297def41014bf3e1aa321a0eaf4d0ebd9855d39a11"

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

export TAG=v0.0.27
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

Verification for index.docker.io/salrashid123/server_image@sha256:120ef883bd7612aee0a100297def41014bf3e1aa321a0eaf4d0ebd9855d39a11 --
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
        "docker-manifest-digest": "sha256:120ef883bd7612aee0a100297def41014bf3e1aa321a0eaf4d0ebd9855d39a11"
      },
      "type": "cosign container image signature"
    },
    "optional": {
      "1.3.6.1.4.1.57264.1.1": "https://token.actions.githubusercontent.com",
      "1.3.6.1.4.1.57264.1.2": "push",
      "1.3.6.1.4.1.57264.1.3": "da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127",
      "1.3.6.1.4.1.57264.1.4": "Release",
      "1.3.6.1.4.1.57264.1.5": "salrashid123/go-bazel-github-workflow",
      "1.3.6.1.4.1.57264.1.6": "refs/tags/v0.0.26",
      "Bundle": {
        "SignedEntryTimestamp": "MEYCIQDZADP2UWtWQfVXVSnb3szVo0iPvqz9bLpzscjRfFgBMgIhANVNk3f6/c8Sfac611nHeoUO+ybIMQvJ4NamW5klmVbJ",
        "Payload": {
          "body": "eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJmYjM5NGZjMWNmZmFmZmY2ZGMwZTFjOGZiMDA2YTQ2NThlNGJiMGFkODE3ZWNmMzVmN2ZkOTVjMTU4NjE0Yzc3In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJUUR0ZUpqWGFDM0ZqYUVEMUQzR2dwZzFMemo0YUszR2ZSMEJMYmVDaDVTVGhBSWdTVndaeEkvTVpDMW9RNmJmdy81b3FnOXBVanQ1ZVpXOVpyUXFXamVVbjFjPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1WRU5EUW5KUFowRjNTVUpCWjBsVlNHUkpObEJWWldwdGJHMVhUV2d5VkRaT1RVUjVSMUJMZW5CRmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUlROTlZFMHhUbFJCTlZkb1kwNU5hbGwzVFdwRk0wMVVVWGRPVkVFMVYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVU0UTJoa09XMVNhblpYS3prd1VESm9hM2xxUjJSUmExSlJWbE5NWlc5VFpXdGplbmdLZW5aQlYybHdZVzh4TWxKTVNHeHpVVGxqZWtGNVJHUkpjRTVYUjFvNGRHdGpiRzV2YWxJMGF6ZExaVGNyUWpodFpEWlBRMEprU1hkbloxaFBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlZaU0hGUUNrcHBhblJHV1ZObVJubG1jbGdyUjNSRWNXOXpSR1JOZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTWsxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRGFHdFpWRkY2V1dwak0xcEhUVEZQVkZwcldXMUZNbGxYUlhsYVJFcHFDazVYUlRKTlJFcHJUakphYVZsdFNteE5WMVY2VFZSSk0wMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUbXBCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFsbDNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJocldWUlJlbGxxWXpNS1drZE5NVTlVV210WmJVVXlXVmRGZVZwRVNtcE9WMFV5VFVSS2EwNHlXbWxaYlVwc1RWZFZlazFVU1ROTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMXBIUlRCTk1ra3pUakpTYWs1VWF6SmFSMHBvVG0xR2FFMXRVWGxaZWxab1RtcEJlVnBFWkcxWmJVcHBXbFJHYkFwTmVrVjVUbnBCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2t5VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxPYWtFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSMUpvVGtST2FRcE9lbVJyV1hwVk5VNXRVbWxaVkZwb1dWUkthMDF0VFRGWlZGbDNUVzFSTTFwdFNtbFpiVlY0V2xSTmVFMXFZM2RHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDVUVlJCZUUxRVozZE9WRUV4VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFphMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNGNqUnBZa0pCUVVGRlFYZENSMDFGVVVOSlFXcFFkSEZaWjJkMk5rRlNSa2d4ZG1nNVVWRk5WVTQyVWxKNlIzWmxiQXBQU25WYVNXWm1ia1ZVVVhaQmFVSm9WVkJxYWs1NFVISXZVbmM1WjNRNE0xZFdNM2xtZUZKbmNsRmxiWGRhY1VkRFpqVk1OM0JUZFhacVFVdENaMmR4Q21ocmFrOVFVVkZFUVhkT2IwRkVRbXhCYWtJclpuSXZiblpEUkVoc1QwWjFiRUZyY0ZJd2RteFFTbmxZWWt4SFUxWlBha040VW1sdVpqVTNSeXN6TldzS1NsRkRTWHBFYjNaV1lrTnNjREZWUlRObVZVTk5VVU5pVUZaR1VIbEVTR3c0U2pKMGVVMUVLekJVYWk5clJrTmtXVUZPU0hOUWMwRXJVbk5uVWl0TlNBcFVjR1JsTVRKalVUZDVUelIzTmt0alRIWTVVMlJSWnowS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19",
          "integratedTime": 1771336509,
          "logIndex": 956940422,
          "logID": "c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d"
        }
      },
      "Issuer": "https://token.actions.githubusercontent.com",
      "Subject": "https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.26",
      "githubWorkflowName": "Release",
      "githubWorkflowRef": "refs/tags/v0.0.26",
      "githubWorkflowRepository": "salrashid123/go-bazel-github-workflow",
      "githubWorkflowSha": "da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127",
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
        "value": "fb394fc1cffafff6dc0e1c8fb006a4658e4bb0ad817ecf35f7fd95c158614c77"
      }
    },
    "signature": {
      "content": "MEUCIQDteJjXaC3FjaED1D3Ggpg1Lzj4aK3GfR0BLbeCh5SThAIgSVwZxI/MZC1oQ6bfw/5oqg9pUjt5eZW9ZrQqWjeUn1c=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMVENDQnJPZ0F3SUJBZ0lVSGRJNlBVZWptbG1XTWgyVDZOTUR5R1BLenBFd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qRTNNVE0xTlRBNVdoY05Nall3TWpFM01UUXdOVEE1V2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUU4Q2hkOW1SanZXKzkwUDJoa3lqR2RRa1JRVlNMZW9TZWtjengKenZBV2lwYW8xMlJMSGxzUTljekF5RGRJcE5XR1o4dGtjbG5valI0azdLZTcrQjhtZDZPQ0JkSXdnZ1hPTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVVZSHFQCkppanRGWVNmRnlmclgrR3REcW9zRGRNd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJMk1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDaGtZVFF6WWpjM1pHTTFPVFprWW1FMllXRXlaREpqCk5XRTJNREprTjJaaVltSmxNV1V6TVRJM01CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5TmpBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNall3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2hrWVRRellqYzMKWkdNMU9UWmtZbUUyWVdFeVpESmpOV0UyTURKa04yWmlZbUpsTVdVek1USTNNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b1pHRTBNMkkzTjJSak5UazJaR0poTm1GaE1tUXlZelZoTmpBeVpEZG1ZbUppWlRGbApNekV5TnpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakkyTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlOakE0QmdvckJnRUVBWU8vTUFFVEJDb01LR1JoTkROaQpOemRrWXpVNU5tUmlZVFpoWVRKa01tTTFZVFl3TW1RM1ptSmlZbVV4WlRNeE1qY3dGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl5TVRBeE1EZ3dOVEExTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZa0dDaXNHQVFRQjFua0NCQUlFZXdSNUFIY0FkUURkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp4cjRpYkJBQUFFQXdCR01FUUNJQWpQdHFZZ2d2NkFSRkgxdmg5UVFNVU42UlJ6R3ZlbApPSnVaSWZmbkVUUXZBaUJoVVBqak54UHIvUnc5Z3Q4M1dWM3lmeFJnclFlbXdacUdDZjVMN3BTdXZqQUtCZ2dxCmhrak9QUVFEQXdOb0FEQmxBakIrZnIvbnZDREhsT0Z1bEFrcFIwdmxQSnlYYkxHU1ZPakN4UmluZjU3RyszNWsKSlFDSXpEb3ZWYkNscDFVRTNmVUNNUUNiUFZGUHlESGw4SjJ0eU1EKzBUai9rRkNkWUFOSHNQc0ErUnNnUitNSApUcGRlMTJjUTd5TzR3NktjTHY5U2RRZz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

the content is the public key

```bash
cat public.crt

-----BEGIN CERTIFICATE-----
MIIHLTCCBrOgAwIBAgIUHdI6PUejmlmWMh2T6NMDyGPKzpEwCgYIKoZIzj0EAwMw
NzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl
cm1lZGlhdGUwHhcNMjYwMjE3MTM1NTA5WhcNMjYwMjE3MTQwNTA5WjAAMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAE8Chd9mRjvW+90P2hkyjGdQkRQVSLeoSekczx
zvAWipao12RLHlsQ9czAyDdIpNWGZ8tkclnojR4k7Ke7+B8md6OCBdIwggXOMA4G
A1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQUYHqP
JijtFYSfFyfrX+GtDqosDdMwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y
ZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy
My9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs
ZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI2MDkGCisGAQQBg78wAQEEK2h0dHBz
Oi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD
vzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBChkYTQzYjc3ZGM1OTZkYmE2YWEyZDJj
NWE2MDJkN2ZiYmJlMWUzMTI3MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB
BAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf
BgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yNjA7BgorBgEEAYO/MAEIBC0M
K2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK
KwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv
LWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl
LnlhbWxAcmVmcy90YWdzL3YwLjAuMjYwOAYKKwYBBAGDvzABCgQqDChkYTQzYjc3
ZGM1OTZkYmE2YWEyZDJjNWE2MDJkN2ZiYmJlMWUzMTI3MB0GCisGAQQBg78wAQsE
DwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi
LmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG
AQQBg78wAQ0EKgwoZGE0M2I3N2RjNTk2ZGJhNmFhMmQyYzVhNjAyZDdmYmJiZTFl
MzEyNzAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI2MBoGCisGAQQB
g78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0
aHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5
BgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv
Z28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh
c2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yNjA4BgorBgEEAYO/MAETBCoMKGRhNDNi
NzdkYzU5NmRiYTZhYTJkMmM1YTYwMmQ3ZmJiYmUxZTMxMjcwFAYKKwYBBAGDvzAB
FAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh
bHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z
LzIyMTAxMDgwNTA1L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw
gYkGCisGAQQB1nkCBAIEewR5AHcAdQDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H
inKALynujgAAAZxr4ibBAAAEAwBGMEQCIAjPtqYggv6ARFH1vh9QQMUN6RRzGvel
OJuZIffnETQvAiBhUPjjNxPr/Rw9gt83WV3yfxRgrQemwZqGCf5L7pSuvjAKBggq
hkjOPQQDAwNoADBlAjB+fr/nvCDHlOFulAkpR0vlPJyXbLGSVOjCxRinf57G+35k
JQCIzDovVbClp1UE3fUCMQCbPVFPyDHl8J2tyMD+0Tj/kFCdYANHsPsA+RsgR+MH
Tpde12cQ7yO4w6KcLv9SdQg=
-----END CERTIFICATE-----
```


which includes

```bash
$  openssl x509 -in public.crt -noout -text


Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            1d:d2:3a:3d:47:a3:9a:59:96:32:1d:93:e8:d3:03:c8:63:ca:ce:91
        Signature Algorithm: ecdsa-with-SHA384
        Issuer: O=sigstore.dev, CN=sigstore-intermediate
        Validity
            Not Before: Feb 17 13:55:09 2026 GMT
            Not After : Feb 17 14:05:09 2026 GMT
        Subject: 
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:f0:28:5d:f6:64:63:bd:6f:bd:d0:fd:a1:93:28:
                    c6:75:09:11:41:54:8b:7a:84:9e:91:cc:f1:ce:f0:
                    16:8a:96:a8:d7:64:4b:1e:5b:10:f5:cc:c0:c8:37:
                    48:a4:d5:86:67:cb:64:72:59:e8:8d:1e:24:ec:a7:
                    bb:f8:1f:26:77
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage: 
                Code Signing
            X509v3 Subject Key Identifier: 
                60:7A:8F:26:28:ED:15:84:9F:17:27:EB:5F:E1:AD:0E:AA:2C:0D:D3
            X509v3 Authority Key Identifier: 
                DF:D3:E9:CF:56:24:11:96:F9:A8:D8:E9:28:55:A2:C6:2E:18:64:3F
            X509v3 Subject Alternative Name: critical
                URI:https://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.26
            1.3.6.1.4.1.57264.1.1: 
                https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.2: 
                push
            1.3.6.1.4.1.57264.1.3: 
                da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127
            1.3.6.1.4.1.57264.1.4: 
                Release
            1.3.6.1.4.1.57264.1.5: 
                salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.6: 
                refs/tags/v0.0.26
            1.3.6.1.4.1.57264.1.8: 
                .+https://token.actions.githubusercontent.com
            1.3.6.1.4.1.57264.1.9: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.26
            1.3.6.1.4.1.57264.1.10: 
                .(da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127
            1.3.6.1.4.1.57264.1.11: 
github-hosted   .
            1.3.6.1.4.1.57264.1.12: 
                .8https://github.com/salrashid123/go-bazel-github-workflow
            1.3.6.1.4.1.57264.1.13: 
                .(da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127
            1.3.6.1.4.1.57264.1.14: 
                ..refs/tags/v0.0.26
            1.3.6.1.4.1.57264.1.15: 
                .
1150755799
            1.3.6.1.4.1.57264.1.16: 
                ..https://github.com/salrashid123
            1.3.6.1.4.1.57264.1.17: 
                ..11149054
            1.3.6.1.4.1.57264.1.18: 
                .ihttps://github.com/salrashid123/go-bazel-github-workflow/.github/workflows/release.yaml@refs/tags/v0.0.26
            1.3.6.1.4.1.57264.1.19: 
                .(da43b77dc596dba6aa2d2c5a602d7fbbbe1e3127
            1.3.6.1.4.1.57264.1.20: 
                ..push
            1.3.6.1.4.1.57264.1.21: 
                .\https://github.com/salrashid123/go-bazel-github-workflow/actions/runs/22101080505/attempts/1
            1.3.6.1.4.1.57264.1.22: 
                ..public
            CT Precertificate SCTs: 
                Signed Certificate Timestamp:
                    Version   : v1 (0x0)
                    Log ID    : DD:3D:30:6A:C6:C7:11:32:63:19:1E:1C:99:67:37:02:
                                A2:4A:5E:B8:DE:3C:AD:FF:87:8A:72:80:2F:29:EE:8E
                    Timestamp : Feb 17 13:55:09.121 2026 GMT
                    Extensions: none
                    Signature : ecdsa-with-SHA256
                                30:44:02:20:08:CF:B6:A6:20:82:FE:80:44:51:F5:BE:
                                1F:50:40:C5:0D:E9:14:73:1A:F7:A5:38:9B:99:21:F7:
                                E7:11:34:2F:02:20:61:50:F8:E3:37:13:EB:FD:1C:3D:
                                82:DF:37:59:5D:F2:7F:14:60:AD:07:A6:C1:9A:86:09:
                                FE:4B:EE:94:AE:BE
    Signature Algorithm: ecdsa-with-SHA384
    Signature Value:
        30:65:02:30:7e:7e:bf:e7:bc:20:c7:94:e1:6e:94:09:29:47:
        4b:e5:3c:9c:97:6c:b1:92:54:e8:c2:c5:18:a7:7f:9e:c6:fb:
        7e:64:25:00:88:cc:3a:2f:55:b0:a5:a7:55:04:dd:f5:02:31:
        00:9b:3d:51:4f:c8:31:e5:f0:9d:ad:c8:c0:fe:d1:38:ff:90:
        50:9d:60:03:47:b0:fb:00:f9:1b:20:47:e3:07:4e:97:5e:d7:
        67:10:ef:23:b8:c3:a2:9c:2e:ff:52:75:08
```


Now you can use `rekor-cli` to inspect the tlog entry

```bash
$ rekor-cli search --rekor_server https://rekor.sigstore.dev    --sha  fb394fc1cffafff6dc0e1c8fb006a4658e4bb0ad817ecf35f7fd95c158614c77
Found matching entries (listed by UUID):
108e9186e8c5677a7dcf8784f72339d46e3a4a04a87c93168066e0d1f1e652f45e76200af31e5bba
```

for our log entry

```bash
$ rekor-cli get --rekor_server https://rekor.sigstore.dev    --uuid 108e9186e8c5677a7dcf8784f72339d46e3a4a04a87c93168066e0d1f1e652f45e76200af31e5bba 

LogID: c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d
Index: 956940422
IntegratedTime: 2026-02-17T13:55:09Z
UUID: 108e9186e8c5677a7dcf8784f72339d46e3a4a04a87c93168066e0d1f1e652f45e76200af31e5bba
Body: {
  "HashedRekordObj": {
    "data": {
      "hash": {
        "algorithm": "sha256",
        "value": "fb394fc1cffafff6dc0e1c8fb006a4658e4bb0ad817ecf35f7fd95c158614c77"
      }
    },
    "signature": {
      "content": "MEUCIQDteJjXaC3FjaED1D3Ggpg1Lzj4aK3GfR0BLbeCh5SThAIgSVwZxI/MZC1oQ6bfw/5oqg9pUjt5eZW9ZrQqWjeUn1c=",
      "publicKey": {
        "content": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUhMVENDQnJPZ0F3SUJBZ0lVSGRJNlBVZWptbG1XTWgyVDZOTUR5R1BLenBFd0NnWUlLb1pJemowRUF3TXcKTnpFVk1CTUdBMVVFQ2hNTWMybG5jM1J2Y21VdVpHVjJNUjR3SEFZRFZRUURFeFZ6YVdkemRHOXlaUzFwYm5SbApjbTFsWkdsaGRHVXdIaGNOTWpZd01qRTNNVE0xTlRBNVdoY05Nall3TWpFM01UUXdOVEE1V2pBQU1Ga3dFd1lICktvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUU4Q2hkOW1SanZXKzkwUDJoa3lqR2RRa1JRVlNMZW9TZWtjengKenZBV2lwYW8xMlJMSGxzUTljekF5RGRJcE5XR1o4dGtjbG5valI0azdLZTcrQjhtZDZPQ0JkSXdnZ1hPTUE0RwpBMVVkRHdFQi93UUVBd0lIZ0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREF6QWRCZ05WSFE0RUZnUVVZSHFQCkppanRGWVNmRnlmclgrR3REcW9zRGRNd0h3WURWUjBqQkJnd0ZvQVUzOVBwejFZa0VaYjVxTmpwS0ZXaXhpNFkKWkQ4d2R3WURWUjBSQVFIL0JHMHdhNFpwYUhSMGNITTZMeTluYVhSb2RXSXVZMjl0TDNOaGJISmhjMmhwWkRFeQpNeTluYnkxaVlYcGxiQzFuYVhSb2RXSXRkMjl5YTJac2IzY3ZMbWRwZEdoMVlpOTNiM0pyWm14dmQzTXZjbVZzClpXRnpaUzU1WVcxc1FISmxabk12ZEdGbmN5OTJNQzR3TGpJMk1Ea0dDaXNHQVFRQmc3OHdBUUVFSzJoMGRIQnoKT2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd0VnWUtLd1lCQkFHRAp2ekFCQWdRRWNIVnphREEyQmdvckJnRUVBWU8vTUFFREJDaGtZVFF6WWpjM1pHTTFPVFprWW1FMllXRXlaREpqCk5XRTJNREprTjJaaVltSmxNV1V6TVRJM01CVUdDaXNHQVFRQmc3OHdBUVFFQjFKbGJHVmhjMlV3TXdZS0t3WUIKQkFHRHZ6QUJCUVFsYzJGc2NtRnphR2xrTVRJekwyZHZMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHpBZgpCZ29yQmdFRUFZTy9NQUVHQkJGeVpXWnpMM1JoWjNNdmRqQXVNQzR5TmpBN0Jnb3JCZ0VFQVlPL01BRUlCQzBNCksyaDBkSEJ6T2k4dmRHOXJaVzR1WVdOMGFXOXVjeTVuYVhSb2RXSjFjMlZ5WTI5dWRHVnVkQzVqYjIwd2VRWUsKS3dZQkJBR0R2ekFCQ1FSckRHbG9kSFJ3Y3pvdkwyZHBkR2gxWWk1amIyMHZjMkZzY21GemFHbGtNVEl6TDJkdgpMV0poZW1Wc0xXZHBkR2gxWWkxM2IzSnJabXh2ZHk4dVoybDBhSFZpTDNkdmNtdG1iRzkzY3k5eVpXeGxZWE5sCkxubGhiV3hBY21WbWN5OTBZV2R6TDNZd0xqQXVNall3T0FZS0t3WUJCQUdEdnpBQkNnUXFEQ2hrWVRRellqYzMKWkdNMU9UWmtZbUUyWVdFeVpESmpOV0UyTURKa04yWmlZbUpsTVdVek1USTNNQjBHQ2lzR0FRUUJnNzh3QVFzRQpEd3dOWjJsMGFIVmlMV2h2YzNSbFpEQklCZ29yQmdFRUFZTy9NQUVNQkRvTU9HaDBkSEJ6T2k4dloybDBhSFZpCkxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNdloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTURnR0Npc0cKQVFRQmc3OHdBUTBFS2d3b1pHRTBNMkkzTjJSak5UazJaR0poTm1GaE1tUXlZelZoTmpBeVpEZG1ZbUppWlRGbApNekV5TnpBaEJnb3JCZ0VFQVlPL01BRU9CQk1NRVhKbFpuTXZkR0ZuY3k5Mk1DNHdMakkyTUJvR0Npc0dBUVFCCmc3OHdBUThFREF3S01URTFNRGMxTlRjNU9UQXZCZ29yQmdFRUFZTy9NQUVRQkNFTUgyaDBkSEJ6T2k4dloybDAKYUhWaUxtTnZiUzl6WVd4eVlYTm9hV1F4TWpNd0dBWUtLd1lCQkFHRHZ6QUJFUVFLREFneE1URTBPVEExTkRCNQpCZ29yQmdFRUFZTy9NQUVTQkdzTWFXaDBkSEJ6T2k4dloybDBhSFZpTG1OdmJTOXpZV3h5WVhOb2FXUXhNak12CloyOHRZbUY2Wld3dFoybDBhSFZpTFhkdmNtdG1iRzkzTHk1bmFYUm9kV0l2ZDI5eWEyWnNiM2R6TDNKbGJHVmgKYzJVdWVXRnRiRUJ5WldaekwzUmhaM012ZGpBdU1DNHlOakE0QmdvckJnRUVBWU8vTUFFVEJDb01LR1JoTkROaQpOemRrWXpVNU5tUmlZVFpoWVRKa01tTTFZVFl3TW1RM1ptSmlZbVV4WlRNeE1qY3dGQVlLS3dZQkJBR0R2ekFCCkZBUUdEQVJ3ZFhOb01Hd0dDaXNHQVFRQmc3OHdBUlVFWGd4Y2FIUjBjSE02THk5bmFYUm9kV0l1WTI5dEwzTmgKYkhKaGMyaHBaREV5TXk5bmJ5MWlZWHBsYkMxbmFYUm9kV0l0ZDI5eWEyWnNiM2N2WVdOMGFXOXVjeTl5ZFc1egpMekl5TVRBeE1EZ3dOVEExTDJGMGRHVnRjSFJ6THpFd0ZnWUtLd1lCQkFHRHZ6QUJGZ1FJREFad2RXSnNhV013CmdZa0dDaXNHQVFRQjFua0NCQUlFZXdSNUFIY0FkUURkUFRCcXhzY1JNbU1aSGh5Wlp6Y0Nva3BldU40OHJmK0gKaW5LQUx5bnVqZ0FBQVp4cjRpYkJBQUFFQXdCR01FUUNJQWpQdHFZZ2d2NkFSRkgxdmg5UVFNVU42UlJ6R3ZlbApPSnVaSWZmbkVUUXZBaUJoVVBqak54UHIvUnc5Z3Q4M1dWM3lmeFJnclFlbXdacUdDZjVMN3BTdXZqQUtCZ2dxCmhrak9QUVFEQXdOb0FEQmxBakIrZnIvbnZDREhsT0Z1bEFrcFIwdmxQSnlYYkxHU1ZPakN4UmluZjU3RyszNWsKSlFDSXpEb3ZWYkNscDFVRTNmVUNNUUNiUFZGUHlESGw4SjJ0eU1EKzBUai9rRkNkWUFOSHNQc0ErUnNnUitNSApUcGRlMTJjUTd5TzR3NktjTHY5U2RRZz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
      }
    }
  }
}
```

you can use crane to inspect the signature too

```bash
$ crane  manifest salrashid123/server_image:sha256-120ef883bd7612aee0a100297def41014bf3e1aa321a0eaf4d0ebd9855d39a11.sig | jq '.'


{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 233,
    "digest": "sha256:66a0e317a7e342e0695af37abbdd40464fe2fbda162b44595121b05dfb72c709"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:fb394fc1cffafff6dc0e1c8fb006a4658e4bb0ad817ecf35f7fd95c158614c77",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEUCIQDteJjXaC3FjaED1D3Ggpg1Lzj4aK3GfR0BLbeCh5SThAIgSVwZxI/MZC1oQ6bfw/5oqg9pUjt5eZW9ZrQqWjeUn1c=",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEYCIQDZADP2UWtWQfVXVSnb3szVo0iPvqz9bLpzscjRfFgBMgIhANVNk3f6/c8Sfac611nHeoUO+ybIMQvJ4NamW5klmVbJ\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiJmYjM5NGZjMWNmZmFmZmY2ZGMwZTFjOGZiMDA2YTQ2NThlNGJiMGFkODE3ZWNmMzVmN2ZkOTVjMTU4NjE0Yzc3In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJUUR0ZUpqWGFDM0ZqYUVEMUQzR2dwZzFMemo0YUszR2ZSMEJMYmVDaDVTVGhBSWdTVndaeEkvTVpDMW9RNmJmdy81b3FnOXBVanQ1ZVpXOVpyUXFXamVVbjFjPSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVaE1WRU5EUW5KUFowRjNTVUpCWjBsVlNHUkpObEJWWldwdGJHMVhUV2d5VkRaT1RVUjVSMUJMZW5CRmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcFpkMDFxUlROTlZFMHhUbFJCTlZkb1kwNU5hbGwzVFdwRk0wMVVVWGRPVkVFMVYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVU0UTJoa09XMVNhblpYS3prd1VESm9hM2xxUjJSUmExSlJWbE5NWlc5VFpXdGplbmdLZW5aQlYybHdZVzh4TWxKTVNHeHpVVGxqZWtGNVJHUkpjRTVYUjFvNGRHdGpiRzV2YWxJMGF6ZExaVGNyUWpodFpEWlBRMEprU1hkbloxaFBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlZaU0hGUUNrcHBhblJHV1ZObVJubG1jbGdyUjNSRWNXOXpSR1JOZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDJSM1dVUldVakJTUVZGSUwwSkhNSGRoTkZwd1lVaFNNR05JVFRaTWVUbHVZVmhTYjJSWFNYVlpNamwwVEROT2FHSklTbWhqTW1od1drUkZlUXBOZVRsdVlua3hhVmxZY0d4aVF6RnVZVmhTYjJSWFNYUmtNamw1WVRKYWMySXpZM1pNYldSd1pFZG9NVmxwT1ROaU0wcHlXbTE0ZG1RelRYWmpiVlp6Q2xwWFJucGFVelUxV1ZjeGMxRklTbXhhYmsxMlpFZEdibU41T1RKTlF6UjNUR3BKTWsxRWEwZERhWE5IUVZGUlFtYzNPSGRCVVVWRlN6Sm9NR1JJUW5vS1QyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDBWbldVdExkMWxDUWtGSFJBcDJla0ZDUVdkUlJXTklWbnBoUkVFeVFtZHZja0puUlVWQldVOHZUVUZGUkVKRGFHdFpWRkY2V1dwak0xcEhUVEZQVkZwcldXMUZNbGxYUlhsYVJFcHFDazVYUlRKTlJFcHJUakphYVZsdFNteE5WMVY2VFZSSk0wMUNWVWREYVhOSFFWRlJRbWMzT0hkQlVWRkZRakZLYkdKSFZtaGpNbFYzVFhkWlMwdDNXVUlLUWtGSFJIWjZRVUpDVVZGc1l6SkdjMk50Um5waFIyeHJUVlJKZWt3eVpIWk1WMHBvWlcxV2MweFhaSEJrUjJneFdXa3hNMkl6U25KYWJYaDJaSHBCWmdwQ1oyOXlRbWRGUlVGWlR5OU5RVVZIUWtKR2VWcFhXbnBNTTFKb1dqTk5kbVJxUVhWTlF6UjVUbXBCTjBKbmIzSkNaMFZGUVZsUEwwMUJSVWxDUXpCTkNrc3lhREJrU0VKNlQyazRkbVJIT1hKYVZ6UjFXVmRPTUdGWE9YVmplVFZ1WVZoU2IyUlhTakZqTWxaNVdUSTVkV1JIVm5Wa1F6VnFZakl3ZDJWUldVc0tTM2RaUWtKQlIwUjJla0ZDUTFGU2NrUkhiRzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmpNa1p6WTIxR2VtRkhiR3ROVkVsNlRESmtkZ3BNVjBwb1pXMVdjMHhYWkhCa1IyZ3hXV2t4TTJJelNuSmFiWGgyWkhrNGRWb3liREJoU0ZacFRETmtkbU50ZEcxaVJ6a3pZM2s1ZVZwWGVHeFpXRTVzQ2t4dWJHaGlWM2hCWTIxV2JXTjVPVEJaVjJSNlRETlpkMHhxUVhWTmFsbDNUMEZaUzB0M1dVSkNRVWRFZG5wQlFrTm5VWEZFUTJocldWUlJlbGxxWXpNS1drZE5NVTlVV210WmJVVXlXVmRGZVZwRVNtcE9WMFV5VFVSS2EwNHlXbWxaYlVwc1RWZFZlazFVU1ROTlFqQkhRMmx6UjBGUlVVSm5OemgzUVZGelJRcEVkM2RPV2pKc01HRklWbWxNVjJoMll6TlNiRnBFUWtsQ1oyOXlRbWRGUlVGWlR5OU5RVVZOUWtSdlRVOUhhREJrU0VKNlQyazRkbG95YkRCaFNGWnBDa3h0VG5aaVV6bDZXVmQ0ZVZsWVRtOWhWMUY0VFdwTmRsb3lPSFJaYlVZMldsZDNkRm95YkRCaFNGWnBURmhrZG1OdGRHMWlSemt6VFVSblIwTnBjMGNLUVZGUlFtYzNPSGRCVVRCRlMyZDNiMXBIUlRCTk1ra3pUakpTYWs1VWF6SmFSMHBvVG0xR2FFMXRVWGxaZWxab1RtcEJlVnBFWkcxWmJVcHBXbFJHYkFwTmVrVjVUbnBCYUVKbmIzSkNaMFZGUVZsUEwwMUJSVTlDUWsxTlJWaEtiRnB1VFhaa1IwWnVZM2s1TWsxRE5IZE1ha2t5VFVKdlIwTnBjMGRCVVZGQ0NtYzNPSGRCVVRoRlJFRjNTMDFVUlRGTlJHTXhUbFJqTlU5VVFYWkNaMjl5UW1kRlJVRlpUeTlOUVVWUlFrTkZUVWd5YURCa1NFSjZUMms0ZGxveWJEQUtZVWhXYVV4dFRuWmlVemw2V1ZkNGVWbFlUbTloVjFGNFRXcE5kMGRCV1V0TGQxbENRa0ZIUkhaNlFVSkZVVkZMUkVGbmVFMVVSVEJQVkVFeFRrUkNOUXBDWjI5eVFtZEZSVUZaVHk5TlFVVlRRa2R6VFdGWGFEQmtTRUo2VDJrNGRsb3liREJoU0ZacFRHMU9kbUpUT1hwWlYzaDVXVmhPYjJGWFVYaE5hazEyQ2xveU9IUlpiVVkyV2xkM2RGb3liREJoU0ZacFRGaGtkbU50ZEcxaVJ6a3pUSGsxYm1GWVVtOWtWMGwyWkRJNWVXRXlXbk5pTTJSNlRETktiR0pIVm1nS1l6SlZkV1ZYUm5SaVJVSjVXbGRhZWt3elVtaGFNMDEyWkdwQmRVMUROSGxPYWtFMFFtZHZja0puUlVWQldVOHZUVUZGVkVKRGIwMUxSMUpvVGtST2FRcE9lbVJyV1hwVk5VNXRVbWxaVkZwb1dWUkthMDF0VFRGWlZGbDNUVzFSTTFwdFNtbFpiVlY0V2xSTmVFMXFZM2RHUVZsTFMzZFpRa0pCUjBSMmVrRkNDa1pCVVVkRVFWSjNaRmhPYjAxSGQwZERhWE5IUVZGUlFtYzNPSGRCVWxWRldHZDRZMkZJVWpCalNFMDJUSGs1Ym1GWVVtOWtWMGwxV1RJNWRFd3pUbWdLWWtoS2FHTXlhSEJhUkVWNVRYazVibUo1TVdsWldIQnNZa014Ym1GWVVtOWtWMGwwWkRJNWVXRXlXbk5pTTJOMldWZE9NR0ZYT1hWamVUbDVaRmMxZWdwTWVrbDVUVlJCZUUxRVozZE9WRUV4VERKR01HUkhWblJqU0ZKNlRIcEZkMFpuV1V0TGQxbENRa0ZIUkhaNlFVSkdaMUZKUkVGYWQyUlhTbk5oVjAxM0NtZFphMGREYVhOSFFWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswZ0thVzVMUVV4NWJuVnFaMEZCUVZwNGNqUnBZa0pCUVVGRlFYZENSMDFGVVVOSlFXcFFkSEZaWjJkMk5rRlNSa2d4ZG1nNVVWRk5WVTQyVWxKNlIzWmxiQXBQU25WYVNXWm1ia1ZVVVhaQmFVSm9WVkJxYWs1NFVISXZVbmM1WjNRNE0xZFdNM2xtZUZKbmNsRmxiWGRhY1VkRFpqVk1OM0JUZFhacVFVdENaMmR4Q21ocmFrOVFVVkZFUVhkT2IwRkVRbXhCYWtJclpuSXZiblpEUkVoc1QwWjFiRUZyY0ZJd2RteFFTbmxZWWt4SFUxWlBha040VW1sdVpqVTNSeXN6TldzS1NsRkRTWHBFYjNaV1lrTnNjREZWUlRObVZVTk5VVU5pVUZaR1VIbEVTR3c0U2pKMGVVMUVLekJVYWk5clJrTmtXVUZPU0hOUWMwRXJVbk5uVWl0TlNBcFVjR1JsTVRKalVUZDVUelIzTmt0alRIWTVVMlJSWnowS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19\",\"integratedTime\":1771336509,\"logIndex\":956940422,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
        "dev.sigstore.cosign/certificate": "-----BEGIN CERTIFICATE-----\nMIIHLTCCBrOgAwIBAgIUHdI6PUejmlmWMh2T6NMDyGPKzpEwCgYIKoZIzj0EAwMw\nNzEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MR4wHAYDVQQDExVzaWdzdG9yZS1pbnRl\ncm1lZGlhdGUwHhcNMjYwMjE3MTM1NTA5WhcNMjYwMjE3MTQwNTA5WjAAMFkwEwYH\nKoZIzj0CAQYIKoZIzj0DAQcDQgAE8Chd9mRjvW+90P2hkyjGdQkRQVSLeoSekczx\nzvAWipao12RLHlsQ9czAyDdIpNWGZ8tkclnojR4k7Ke7+B8md6OCBdIwggXOMA4G\nA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAzAdBgNVHQ4EFgQUYHqP\nJijtFYSfFyfrX+GtDqosDdMwHwYDVR0jBBgwFoAU39Ppz1YkEZb5qNjpKFWixi4Y\nZD8wdwYDVR0RAQH/BG0wa4ZpaHR0cHM6Ly9naXRodWIuY29tL3NhbHJhc2hpZDEy\nMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvLmdpdGh1Yi93b3JrZmxvd3MvcmVs\nZWFzZS55YW1sQHJlZnMvdGFncy92MC4wLjI2MDkGCisGAQQBg78wAQEEK2h0dHBz\nOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20wEgYKKwYBBAGD\nvzABAgQEcHVzaDA2BgorBgEEAYO/MAEDBChkYTQzYjc3ZGM1OTZkYmE2YWEyZDJj\nNWE2MDJkN2ZiYmJlMWUzMTI3MBUGCisGAQQBg78wAQQEB1JlbGVhc2UwMwYKKwYB\nBAGDvzABBQQlc2FscmFzaGlkMTIzL2dvLWJhemVsLWdpdGh1Yi13b3JrZmxvdzAf\nBgorBgEEAYO/MAEGBBFyZWZzL3RhZ3MvdjAuMC4yNjA7BgorBgEEAYO/MAEIBC0M\nK2h0dHBzOi8vdG9rZW4uYWN0aW9ucy5naXRodWJ1c2VyY29udGVudC5jb20weQYK\nKwYBBAGDvzABCQRrDGlodHRwczovL2dpdGh1Yi5jb20vc2FscmFzaGlkMTIzL2dv\nLWJhemVsLWdpdGh1Yi13b3JrZmxvdy8uZ2l0aHViL3dvcmtmbG93cy9yZWxlYXNl\nLnlhbWxAcmVmcy90YWdzL3YwLjAuMjYwOAYKKwYBBAGDvzABCgQqDChkYTQzYjc3\nZGM1OTZkYmE2YWEyZDJjNWE2MDJkN2ZiYmJlMWUzMTI3MB0GCisGAQQBg78wAQsE\nDwwNZ2l0aHViLWhvc3RlZDBIBgorBgEEAYO/MAEMBDoMOGh0dHBzOi8vZ2l0aHVi\nLmNvbS9zYWxyYXNoaWQxMjMvZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93MDgGCisG\nAQQBg78wAQ0EKgwoZGE0M2I3N2RjNTk2ZGJhNmFhMmQyYzVhNjAyZDdmYmJiZTFl\nMzEyNzAhBgorBgEEAYO/MAEOBBMMEXJlZnMvdGFncy92MC4wLjI2MBoGCisGAQQB\ng78wAQ8EDAwKMTE1MDc1NTc5OTAvBgorBgEEAYO/MAEQBCEMH2h0dHBzOi8vZ2l0\naHViLmNvbS9zYWxyYXNoaWQxMjMwGAYKKwYBBAGDvzABEQQKDAgxMTE0OTA1NDB5\nBgorBgEEAYO/MAESBGsMaWh0dHBzOi8vZ2l0aHViLmNvbS9zYWxyYXNoaWQxMjMv\nZ28tYmF6ZWwtZ2l0aHViLXdvcmtmbG93Ly5naXRodWIvd29ya2Zsb3dzL3JlbGVh\nc2UueWFtbEByZWZzL3RhZ3MvdjAuMC4yNjA4BgorBgEEAYO/MAETBCoMKGRhNDNi\nNzdkYzU5NmRiYTZhYTJkMmM1YTYwMmQ3ZmJiYmUxZTMxMjcwFAYKKwYBBAGDvzAB\nFAQGDARwdXNoMGwGCisGAQQBg78wARUEXgxcaHR0cHM6Ly9naXRodWIuY29tL3Nh\nbHJhc2hpZDEyMy9nby1iYXplbC1naXRodWItd29ya2Zsb3cvYWN0aW9ucy9ydW5z\nLzIyMTAxMDgwNTA1L2F0dGVtcHRzLzEwFgYKKwYBBAGDvzABFgQIDAZwdWJsaWMw\ngYkGCisGAQQB1nkCBAIEewR5AHcAdQDdPTBqxscRMmMZHhyZZzcCokpeuN48rf+H\ninKALynujgAAAZxr4ibBAAAEAwBGMEQCIAjPtqYggv6ARFH1vh9QQMUN6RRzGvel\nOJuZIffnETQvAiBhUPjjNxPr/Rw9gt83WV3yfxRgrQemwZqGCf5L7pSuvjAKBggq\nhkjOPQQDAwNoADBlAjB+fr/nvCDHlOFulAkpR0vlPJyXbLGSVOjCxRinf57G+35k\nJQCIzDovVbClp1UE3fUCMQCbPVFPyDHl8J2tyMD+0Tj/kFCdYANHsPsA+RsgR+MH\nTpde12cQ7yO4w6KcLv9SdQg=\n-----END CERTIFICATE-----\n",
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
      "digest": "lttqqAhC62a1pK9vyX3tnaxp2RHLCoXrYVLYD1G1rWU="
    },
    "signature": "HE2Dqv/GjoH0eS/O4R+T+xauvAOMiRrag/cn3NZTHIvveQG2qoNQg8gzWs4s1U2yFPpgxdaLInS3LXu9aQ9ByQBtBZnHdFsNxfaZSEsauOa79EJGfVLzyNOSiVPT9UG0AkvsEaqVhYdm1u5hZUUsAl8GpThoMrTzL5optdCrK05FIdpKOgoTccEr0hx7JhHBtLA9ox98NjJLWpRJ6ItaDmnzMNMAVLhLmYJpQ1t9hgxQwrRY8b8+Lv6BsvRiG6EXoTktqpPGXLO/jxKI96Hr6fYmXCysrbu5biIiREzOzn8M69PtNJJsp4guir9mvSCPzhDnYYn6++Or99OwgX8duA=="
  }
}


### which can be verified locally with the local signing key
export sig=`cat bazel-bin/app/server_linux_amd64.sig | jq -r '.messageSignature.signature'`
cosign verify-blob --key certs/import-cosign.pub --signature $sig bazel-bin/app/server_linux_amd64_bin
```