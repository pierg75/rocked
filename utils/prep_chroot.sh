#!/usr/bin/env bash



TMPDIR="/tmp"
TMPEXTRACT=$(mktemp -d -t chroot_XXXXXX)
FEDORAIMAGE=Fedora-Container-Minimal-Base-39-1.5.x86_64.tar.xz
CHROOTFILE=Fedora-minimal-chroot.tar


if ! [[ -f "${TMPDIR}/${FEDORAIMAGE}" ]]; then
        wget -P "${TMPDIR}" https://ask4.mm.fcix.net/fedora/linux/releases/39/Container/x86_64/images/"${FEDORAIMAGE}"
fi
tar xvf "${TMPDIR}/${FEDORAIMAGE}" -C "${TMPEXTRACT}"

EXTRACTDIR=$(find "${TMPEXTRACT}" -type d| tail -1)
FILETOCOPY=$(find "${EXTRACTDIR}" -name "*.tar")

mv "${FILETOCOPY}" "${CHROOTFILE}"
RET=$?
if [[ $RET -eq 0 ]]; then
        rm -rf "${TMPEXTRACT}"
fi
echo "The file "${CHROOTFILE}" contains the chroot environment."
