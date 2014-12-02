FROM progrium/busybox

RUN opkg-install ca-certificates

RUN for cert in `ls -1 /etc/ssl/certs/*.crt | grep -v /etc/ssl/certs/ca-certificates.crt`; \
      do cat "$cert" >> /etc/ssl/certs/ca-certificates.crt; \
    done

ADD cf /usr/bin/cf
ADD /tmp/autopilot /usr/bin/autopilot
RUN /usr/bin/cf install-plugin /usr/bin/autopilot

ADD built-check /opt/resource/check
ADD built-out /opt/resource/out
