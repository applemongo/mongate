package server

import (
    "github.com/applemongo/mongate/gate"
    "github.com/Sirupsen/logrus"

    "os"
    "encoding/json"
    "fmt"
    "path/filepath"
    "log"
    "net"
    "errors"
)

type ServerConfig struct {
    BindIp net.IP   `json:"bind_ip,omitempty"`
    BindPort int    `json:"bind_port,omitempty"`
}
type Config struct {
    Server ServerConfig `json:"server"`
    Proxies map[string]gate.Backend `json:"proxies"`
}

const DefaultConfigFile string = "mongate.json"

func NewConfig(logger *logrus.Logger, path string) (Config, error) {
    var config Config
    if path == "" {
        dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
        if err != nil {
            log.Fatal(err)
            log.Fatal("the current dir calculate error: ", err)
            return config, err
        }
        path = dir + "/" + DefaultConfigFile
    }
    fmt.Println(path)

    file, err := os.Open(path)
    if err != nil {
        log.Fatal("configuration file open error: ", err)
        return config, err
    }

    err = json.NewDecoder(file).Decode(&config)
    if err != nil {
        log.Fatal("configuration json parse error: ", err)
        return config, err
    }
    fmt.Println(config)
    serverConfig := &config.Server
    if serverConfig.BindIp == nil {
        return config, errors.New("server.bind_ip is required, e.g. \"127.0.0.1\"")
    }

    if serverConfig.BindPort == nil {
        return config, errors.New("server.bind_port is required, e.g. \"12305\"")
    }

    for id, proxy := range config.Proxies {
        if id == nil || len(id) == 0 {
            log.Fatal("proxies.id is required, e.g. \"mongodb\"")
            return config, errors.New("proxy config error")
        }

        if proxy.BindIP == nil {
            log.Fatal("proxies[id].bind_ip is required, e.g. \"127.0.0.1\"")
            return config, errors.New("proxy config error")
        }

        if proxy.BindPort == nil {
            log.Fatal("proxies[id].bind_ip is required, e.g. \"27016\"")
            return config, errors.New("proxy config error")
        }

        if proxy.IP == nil {
            log.Fatal("proxies[id].ip is required, e.g. \"127.0.0.1\"")
            return config, errors.New("proxy config error")
        }

        if proxy.Port == nil {
            log.Fatal("proxies[id].bind_ip is required, e.g. \"27017\"")
            return config, errors.New("proxy config error")
        }

        if proxy.Name == nil || len(proxy.Name) == 0 {
            proxy.Name = id
        }

        if proxy.Proto == nil || len(proxy.Proto) == 0 {
            proxy.Proto = "tcp"
        }
    }

    fmt.Println(config)
    //fmt.Println("bind ip:\t", config.Server.BindIp)
    //fmt.Println("bind port:\t", config.Server.BindPort)
    //mongodb := config.Proxies["mongodb"]
    //fmt.Println("proxies - mongodb bind ip:\t", mongodb.BindIP)
    //fmt.Println("proxies - mongodb bind port:\t", mongodb.BindPort)
    //fmt.Println("proxies - mongodb ip:\t", mongodb.IP)
    //fmt.Println("proxies - mongodb port:\t", mongodb.Port)

    return config, nil
}
