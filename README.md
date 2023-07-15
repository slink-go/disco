Disco is a discovery service (a kind of [eureka](https://github.com/Netflix/eureka) alternative) written in go.

The simplest way to run this service is to use docker compose:

```yaml
services:
  disco:
    image: slinkgo/disco:alpine
    container_name: disco
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - LOGGING_LEVEL=INFO
```

TODO: 
- java client
  - plain java
  - spring boot starter
- handle port-only endpoints from clients (remote-ip, X-CLIENT-IP, etc)
