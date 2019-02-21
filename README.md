Bamboo basket
====
*批量压缩文件*

### 参数
go version go1.11+
- -s  压缩的文件目录，默认为./source
- -o  导出目录，默认为./out
- -k  设置压缩文件密码，默认无密码
- -r  开启文件随机重命名，默认不开启 (true、false)

### build
```
$ go build basket.go
$ ./basket -s source -k password -r true
```