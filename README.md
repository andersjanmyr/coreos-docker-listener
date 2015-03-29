# Docker Etcd Bridge
Stores Docker container information in Etcd when a new container is
started.

The information is registered as

```
/docker/machines/<machine-name>
/docker/machines/<machine-name>/awsinfo
/docker/machines/<machine-name>/containers/<container-id>
```

The information stored comes from `docker inspect <container-id>`.


