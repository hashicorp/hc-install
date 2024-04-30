# Contributing Notes

## Releasing

Releases are made on a reasonably regular basis by the maintainers (HashiCorp staff), using our internal tooling. The following notes are only relevant to maintainers.

Release process:

1. Update [`version/VERSION`](https://github.com/hashicorp/hc-install/blob/main/version/VERSION) to remove `-dev` suffix and set it to the intended version to be released
1. Wait for [`build` workflow](https://github.com/hashicorp/hc-install/actions/workflows/build.yml) to finish
1. Run the Release workflow with the appropriate version (matching the one in `version/VERSION`) & SHA (long one).
1. Wait for a message in the Slack channel saying that authorisation is needed to promote artifacts to production. Click on the link and approve.
1. Once notified that promotion is successful, go to <https://github.com/hashicorp/crt-workflows-common/actions/workflows/promote-production-packaging.yml>, locate the hc-install promote-production-packaging workflow, and approve.
