FROM gliderlabs/alpine

RUN apk-install ca-certificates

ADD cf /usr/bin/cf
ADD autopilot /usr/bin/autopilot
RUN cf install-plugin /usr/bin/autopilot

ADD built-check /opt/resource/check
ADD built-out /opt/resource/out
