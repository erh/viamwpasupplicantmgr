
manages wpa supplicant entries
===

Sample Config 
====
```
{
  "filename" : "/etc/wpa_supplicant/wpa_supplicant.conf",
  "networks": [
    {
      "PSK": "bar",
      "SSID": "foo"
    },
    {
      "PSK": "xxx",
      "SSID": "asdljasldwoioisafoiahsoifhasodfas",
      "Encoded" : true
    }

  ]
}
```

to compile for arm64
====
```
env GOOS=linux GOARCH=arm64 make module
viam module upload --platform "linux/arm64" --version <FILL ME IN> module.tar.gz
```
