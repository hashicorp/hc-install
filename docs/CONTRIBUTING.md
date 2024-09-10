# Contributing Notes

## Releasing

Releases are made on a reasonably regular basis by the maintainers (HashiCorp staff), using our internal tooling. The following notes are only relevant to maintainers.

Release process:

1. Update [`version/VERSION`](https://github.com/hashicorp/hc-install/blob/main/version/VERSION) to remove `-dev` suffix and set it to the intended version to be released
1. Wait for [`build` workflow](https://github.com/hashicorp/hc-install/actions/workflows/build.yml) to finish
1. Run the Release workflow with the appropriate version (matching the one in `version/VERSION`) & SHA (long one).
1. Wait for a message in the Slack channel saying that authorisation is needed to promote artifacts to production. Click on the link and approve.
1. Once notified that promotion is successful, go to <https://github.com/hashicorp/crt-workflows-common/actions/workflows/promote-production-packaging.yml>, locate the hc-install promote-production-packaging workflow, and approve.
1. Update [`version/VERSION`](https://github.com/hashicorp/hc-install/blob/main/version/VERSION) and add `-dev` suffix and set it to the expected next version to be released

### Obtaining perms to do a release

The process to obtain the permissions and accesses required in order to do a release for this project has been streamlined and no longer requires a series of helpdesk tickets. Just request membership in [this Passport group](https://passport.hashicorp.services/namespaces/doormat/groups/github-hashicorp-team-release-approvers-hc-install-oss); click the "Add membership" button and add yourself in the format `users/<youremail>`. "Permanent" access is fine. Anyone who is already a member of the releasers group can approve your request to join.
