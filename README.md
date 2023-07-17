Disco is a discovery service (a kind of [eureka](https://github.com/Netflix/eureka) alternative) written in go.

The simplest way to run this service is to use docker compose ([see here](https://hub.docker.com/repository/docker/slinkgo/disco/general)):

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
- simplify client endpoints registration: handle port-only endpoints from clients
  (remote-ip, X-Forwarded-For / X-Real-IP / X-CLIENT-IP, etc)
- implement redis backend
- implement etcd backend
- implement multinode-consensus backend
- fix Let's Encrypt support 
