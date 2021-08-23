FROM python:buster

RUN apt-get -y update && \
    apt-get install -y ffmpeg bash python3 python3-pip &&\
    apt-get -y update && \
    apt-get clean all && \
    python3 -m pip install --upgrade git+https://github.com/yt-dlp/yt-dlp.git@release && \
    python3 -m pip install apprise

ADD vd_agent /app/vd_agent

CMD [ "/app/vd_agent", "/app/config.yaml" ]