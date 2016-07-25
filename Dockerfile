FROM concourse/buildroot:base

ADD cf /usr/bin/cf

ADD autopilot /usr/bin/autopilot
RUN /usr/bin/cf install-plugin -f /usr/bin/autopilot

ADD antifreeze /usr/bin/antifreeze
RUN /usr/bin/cf install-plugin -f /usr/bin/antifreeze

ADD built-check /opt/resource/check
ADD built-out /opt/resource/out
ADD built-in /opt/resource/in
