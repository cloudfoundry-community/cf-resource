FROM progrium/busybox

RUN opkg-install ca-certificates

# satisfy go crypto/x509
RUN bash -c "cat /etc/ssl/certs/*.pem > /etc/ssl/certs/ca-certificates.crt"

ADD cf /usr/bin/cf
ADD autopilot /usr/bin/autopilot
RUN /usr/bin/cf install-plugin /usr/bin/autopilot

ADD built-check /opt/resource/check
ADD built-out /opt/resource/out
