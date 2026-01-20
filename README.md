# Rainbow

```shell 安装
docker run -d --name rainbow-server  --restart=always -v /data:/data \
 -v /etc/localtime:/etc/localtime:ro  \
 -v /root/.ssh:/root/.ssh \
 -v /usr/bin:/usr/bin -v /var/run/docker.sock:/var/run/docker.sock  \
 --network host swr.cn-north-4.myhuaweicloud.com/pixiu-public/rainbowd:v1 /data/app --configFile /data/config.yaml
```

Copyright 2019 caoyingjun (cao.yingjunz@gmail.com) Apache License 2.0
