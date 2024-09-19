## 打包

```shell
make
make linux
make arm
```

## 映射

1. -m

    ```shell
    ./tproxy -l 192.168.136.150 -m "11111>192.168.218.52:54321"
    
    ./tproxy -l 192.168.136.150 -m "58088>192.168.218.52:5236,54321>192.168.218.52:54321"
    
    nohup ./tproxy -l 192.168.136.150 -m "58088>192.168.218.52:5236,54321>192.168.218.52:54321" > proxy.log 2>&1 &
    ```

2. -mf

    ```shell
    cat << EOF > proxy.mapping
   58088>192.168.218.52:5236
   54321>192.168.218.52:54321
   EOF
    ./tproxy -l 192.168.136.150 -mf proxy.mapping
   nohup ./tproxy -l 192.168.136.150 -mf proxy.mapping  > proxy.log 2>&1 &
    ```

## 测试

```shell
telnet 192.168.136.150 11111
```