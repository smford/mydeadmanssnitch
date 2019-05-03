FROM alpine:3.7
RUN apk add --no-cache curl
RUN curl -O -L https://github.com/smford/snitchit/raw/master/binaries/x86_64/snitchit
RUN chmod +x /snitchit 
COPY files/config.yaml /config.yaml
RUN cat /config.yaml
RUN /snitchit --message insidealpine1
