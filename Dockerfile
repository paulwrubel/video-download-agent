# FROM debian:bullseye-slim

# ENV LC_ALL=C.UTF-8 \
#     LANG=C.UTF-8 \
#     LANGUAGE=en_US:en

# SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# RUN set -x && \
#     TEMP_PACKAGES=() && \
#     KEPT_PACKAGES=() && \
#     # Packages only required during build
#     TEMP_PACKAGES+=(git) && \
#     TEMP_PACKAGES+=(make) && \
#     TEMP_PACKAGES+=(pandoc) && \
#     # Packages kept in the image
#     KEPT_PACKAGES+=(bash) && \
#     TEMP_PACKAGES+=(build-essential) && \
#     KEPT_PACKAGES+=(ca-certificates) && \
#     KEPT_PACKAGES+=(ffmpeg) && \
#     KEPT_PACKAGES+=(locales) && \
#     KEPT_PACKAGES+=(locales-all) && \
#     KEPT_PACKAGES+=(mpv) && \
#     KEPT_PACKAGES+=(python3) && \
#     TEMP_PACKAGES+=(python3-dev) && \
#     KEPT_PACKAGES+=(python-is-python3) && \
#     TEMP_PACKAGES+=(python3-pip) && \
#     KEPT_PACKAGES+=(rtmpdump) && \
#     KEPT_PACKAGES+=(zip) && \
#     KEPT_PACKAGES+=(atomicparsley) && \
#     KEPT_PACKAGES+=(aria2) && \
#     KEPT_PACKAGES+=(libdbus-glib-1-dev) && \
#     # Install packages
#     apt-get update -y && \
#     apt-get install -y --no-install-recommends \
#     ${KEPT_PACKAGES[@]} \
#     ${TEMP_PACKAGES[@]} \
#     && \
#     git config --global advice.detachedHead false && \
#     # Install required python modules
#     python3 -m pip install -U pip && \
#     python3 -m pip install --no-cache-dir \
#     pyxattr \
#     pycryptodome \
#     websockets \
#     secretstorage \
#     dbus-python \
#     keyring \
#     mutagen \
#     && \
#     # Install yt-dlp
#     git clone https://github.com/yt-dlp/yt-dlp.git /src/yt-dlp && \
#     pushd /src/yt-dlp && \
#     BRANCH_YTDLP=$(git tag --sort='creatordate' | tail -1) && \
#     git checkout "${BRANCH_YTDLP}" && \
#     export DESTDIR="" \
#     make && \
#     make install && \
#     popd && \
#     # Create /config directory
#     mkdir -p /config && \
#     # Clean-up.
#     apt-get remove -y ${TEMP_PACKAGES[@]} && \
#     apt-get autoremove -y && \
#     apt-get clean -y && \
#     rm -rf /var/lib/apt/lists/* /tmp/* /src && \
#     yt-dlp --version > /CONTAINER_VERSION


# # FROM python:buster

# # RUN apt-get -y update && \
# #     apt-get install -y ffmpeg bash python3 python3-pip &&\
# #     apt-get -y update && \
# #     apt-get clean all && \
# #     python3 -m pip install --upgrade git+https://github.com/yt-dlp/yt-dlp.git@release

FROM python:3-bullseye

RUN set -x && \
    apt-get update -y && \
    apt-get install -y --no-install-recommends \
    ffmpeg && \
    python3 -m pip install -U pip && \
    python3 -m pip install --no-cache-dir \
    yt_dlp \
    pyyaml \
    apscheduler

ADD vd_agent.py /app/vd_agent.py

CMD [ "python3", "-u", "/app/vd_agent.py" ]