# Copyright IBM Corp. 2020, 2026
# SPDX-License-Identifier: MPL-2.0

schema = 1
artifacts {
  zip = [
    "hc-install_${version}_darwin_amd64.zip",
    "hc-install_${version}_darwin_arm64.zip",
    "hc-install_${version}_freebsd_386.zip",
    "hc-install_${version}_freebsd_amd64.zip",
    "hc-install_${version}_freebsd_arm.zip",
    "hc-install_${version}_freebsd_arm64.zip",
    "hc-install_${version}_linux_386.zip",
    "hc-install_${version}_linux_amd64.zip",
    "hc-install_${version}_linux_arm.zip",
    "hc-install_${version}_linux_arm64.zip",
    "hc-install_${version}_openbsd_386.zip",
    "hc-install_${version}_openbsd_amd64.zip",
    "hc-install_${version}_solaris_amd64.zip",
    "hc-install_${version}_windows_386.zip",
    "hc-install_${version}_windows_amd64.zip",
    "hc-install_${version}_windows_arm64.zip",
  ]
  rpm = [
    "hc-install-${version_linux}-1.aarch64.rpm",
    "hc-install-${version_linux}-1.armv7hl.rpm",
    "hc-install-${version_linux}-1.i386.rpm",
    "hc-install-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "hc-install_${version_linux}-1_amd64.deb",
    "hc-install_${version_linux}-1_arm64.deb",
    "hc-install_${version_linux}-1_armhf.deb",
    "hc-install_${version_linux}-1_i386.deb",
  ]
}
