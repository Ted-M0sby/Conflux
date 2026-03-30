package router

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// BuildNacosConfigClient creates a Nacos config client.
func BuildNacosConfigClient(namespaceID, username, password, logDir, cacheDir string, notLoadCache bool, serverHostPorts []string) (config_client.IConfigClient, error) {
	sc, err := buildServerConfigs(serverHostPorts)
	if err != nil {
		return nil, err
	}
	cc := constant.ClientConfig{
		NamespaceId:         namespaceID,
		Username:            username,
		Password:            password,
		LogDir:              logDir,
		CacheDir:            cacheDir,
		NotLoadCacheAtStart: notLoadCache,
		TimeoutMs:           10_000,
	}
	return clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
}

func buildServerConfigs(hostports []string) ([]constant.ServerConfig, error) {
	if len(hostports) == 0 {
		return nil, fmt.Errorf("no nacos server hosts")
	}
	out := make([]constant.ServerConfig, 0, len(hostports))
	for _, hp := range hostports {
		hp = strings.TrimSpace(hp)
		if hp == "" {
			continue
		}
		host, portStr, err := netSplitHostPort(hp)
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseUint(portStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid port in %q", hp)
		}
		out = append(out, constant.ServerConfig{
			IpAddr:      host,
			Port:        port,
			ContextPath: constant.DEFAULT_CONTEXT_PATH,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid nacos server hosts")
	}
	return out, nil
}

func netSplitHostPort(hp string) (host, port string, err error) {
	// Expect host:port; IPv6 [addr]:port not required for demo compose.
	i := strings.LastIndex(hp, ":")
	if i <= 0 || i == len(hp)-1 {
		return "", "", fmt.Errorf("invalid host:port %q", hp)
	}
	return hp[:i], hp[i+1:], nil
}

// NacosLoader wires Nacos config to route table updates.
type NacosLoader struct {
	Client config_client.IConfigClient
}

// StartListen loads initial config and subscribes to changes.
func (l *NacosLoader) StartListen(dataID, group string, onTable func(*Table) error) error {
	content, err := l.Client.GetConfig(vo.ConfigParam{DataId: dataID, Group: group})
	if err != nil {
		return err
	}
	if err := applyNacosYAML([]byte(content), onTable); err != nil {
		return err
	}
	return l.Client.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			_ = applyNacosYAML([]byte(data), onTable)
		},
	})
}

func applyNacosYAML(data []byte, onTable func(*Table) error) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	t, err := ParseYAML(data)
	if err != nil {
		return err
	}
	if len(t.Routes()) == 0 {
		return nil
	}
	if onTable == nil {
		return nil
	}
	return onTable(t)
}
