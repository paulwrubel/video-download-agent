FROM alpine:3.14

RUN apk add --no-cache bash ffmpeg python3 py-pip zip py3-pycryptodome mutagen coreutils
RUN python3 -m pip install --no-cache-dir -U yt-dlp

ADD vd_agent /app/vd_agent

CMD [ "/app/vd_agent", "/app/config.yaml" ]