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

func NewConfig(logger *logrus.Logger, path string) (*Config, error) {
    var config Config
    if path == "" {
        dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
        if err != nil {
            log.Fatal(err)
            log.Fatal("the current dir calculate error: ", err)
            return &config, err
        }
        path = dir + "/" + DefaultConfigFile
    }
    fmt.Println("use config file:", path)

    file, err := os.Open(path)
    if err != nil {
        log.Fatal("configuration file open error: ", err)
        return &config, err
    }

    err = json.NewDecoder(file).Decode(&config)
    if err != nil {
        log.Fatal("configuration json parse error: ", err)
        return &config, err
    }
    fmt.Println("use config: ", config)

    if config.Server.BindIp == nil {
        return &config, errors.New("server.bind_ip is required, e.g. \"127.0.0.1\"")
    }

    for id, proxy := range config.Proxies {
        if len(id) == 0 {
            log.Fatal("proxies.id is required, e.g. \"mongodb\"")
            return &config, errors.New("proxy config error")
        }

        if proxy.BindIP == nil {
            log.Fatal("proxies[id].bind_ip is required, e.g. \"127.0.0.1\"")
            return &config, errors.New("proxy config error")
        }

        if proxy.IP == nil {
            log.Fatal("proxies[id].ip is required, e.g. \"127.0.0.1\"")
            return &config, errors.New("proxy config error")
        }

        if len(proxy.Name) == 0 {
            proxy.Name = id
        }

        if len(proxy.Proto) == 0 {
            proxy.Proto = "tcp"
        }

        fmt.Printf("proxies - %s\n============================\n", id)
        fmt.Printf("proxies - %s - protocol:\t\t%s\n", id, proxy.Proto)
        fmt.Printf("proxies - %s - bind ip:\t\t%s\n", id, proxy.BindIP)
        fmt.Printf("proxies - %s - bind port:\t%v\n", id, proxy.BindPort)
        fmt.Printf("proxies - %s - ip:\t\t\t%s\n", id, proxy.IP)
        fmt.Printf("proxies - %s - port:\t\t\t%v\n", id, proxy.Port)
        fmt.Printf("proxies - %s - connection buffer:\t%v\n", id, proxy.ConnectionBuffer)
        fmt.Printf("proxies - %s - max concurrent:\t\t%v\n", id, proxy.MaxConcurrent)
        fmt.Printf("\n")

        config.Proxies[id] = proxy
    }

    return &config, nil
}
