#!/usr/bin/env bash

TMPDIR="/tmp"
TMPEXTRACT=$(mktemp -d -t chroot_XXXXXX)

declare -A distros
distros[fedora]=https://ask4.mm.fcix.net/fedora/linux/releases/39/Container/x86_64/images/Fedora-Container-Minimal-Base-39-1.5.x86_64.tar.xz
distros[ubuntu]=https://cloud-images.ubuntu.com/releases/server/noble/release/ubuntu-24.04-server-cloudimg-amd64-root.tar.xz


function usage() {
        echo "$0 [fedora|ubuntu]"
        exit 1
}

# Fedora images needs to be extracted first
function fedora() {
        CHROOTFILE=Fedora-minimal-chroot.tar
        if ! [[ -f "${TMPDIR}/${IMAGENAME}" ]]; then
                wget -P "${TMPDIR}" "${distros[$IMAGE]}"
        fi

        tar xvf "${TMPDIR}/${IMAGENAME}" -C "${TMPEXTRACT}"

        EXTRACTDIR=$(find "${TMPEXTRACT}" -type d| tail -1)
        FILETOCOPY=$(find "${EXTRACTDIR}" -name "*.tar")

        mv "${FILETOCOPY}" "${CHROOTFILE}"
        RET=$?
        if [[ $RET -eq 0 ]]; then
                rm -rf "${TMPEXTRACT}"
        fi
        echo "The file "${CHROOTFILE}" contains the chroot environment."
}


# Ubuntu images are already just the environment, so no need to find the right file
# to extract
function ubuntu() {
        CHROOTFILE=Ubuntu-minimal-chroot.tar
        if ! [[ -f "${TMPDIR}/${IMAGENAME}" ]]; then
                wget -P "${TMPDIR}" "${distros[$IMAGE]}"
        fi

        cp "${TMPDIR}/${IMAGENAME}" "${CHROOTFILE}"
        echo "The file "${CHROOTFILE}" contains the chroot environment."
}



if [[ $# -ne 1 ]]; then
        usage
fi

IMAGE=$1
if [[ -v "distros[$IMAGE]" ]]; then
        IMAGENAME=$(basename "${distros[$IMAGE]}")
else
        usage
fi

$IMAGE
