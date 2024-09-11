# Contributing Notes

## Releasing

Releases are made on a reasonably regular basis by the maintainers (HashiCorp staff), using our internal tooling. The following notes are only relevant to maintainers.

Release process:

1. Update [`version/VERSION`](https://github.com/hashicorp/hc-install/blob/main/version/VERSION) to remove `-dev` suffix and set it to the intended version to be released
1. Wait for [`build` workflow](https://github.com/hashicorp/hc-install/actions/workflows/build.yml) to finish
1. Wait for [`prepare` workflow](https://github.com/hashicorp/crt-workflows-common/actions/workflows/crt-prepare.yml) to finish
1. Run the [`release` workflow](https://github.com/hashicorp/hc-install/actions/workflows/release.yml) with the appropriate version (matching the one in `version/VERSION`) & SHA (long one).
   - This will kick off staging and then production promotions. However, there is currently no way of blocking the production promotion until (staging) artifacts are in place, so **the production promotion is likely to fail**.
1. Wait for [`promote-staging` workflow](https://github.com/hashicorp/crt-workflows-common/actions/workflows/crt-promote-staging.yml) to finish.
1. Retry the failed `production` job of the `release` workflow
1. Wait for a message [in the Slack channel](https://hashicorp.enterprise.slack.com/archives/C01QDH3Q37W) saying that authorisation is needed to promote artifacts to production. Click on the link and approve.
1. Wait for the [`promote-production` workflow](https://github.com/hashicorp/crt-workflows-common/actions/workflows/crt-promote-production.yml) to finish
1. Click on the pencil icon of [the latest release](https://github.com/hashicorp/hc-install/releases), click `Generate release notes`, make any manual changes if necessary (e.g. to call out any breaking changes) and `Update release`.
1. Update [`version/VERSION`](https://github.com/hashicorp/hc-install/blob/main/version/VERSION) and add `-dev` suffix and set it to the expected next version to be released
