package utilkubernetes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"

	utilfile "github.com/mjudeikis/portal/pkg/util/file"
)

func GetRestConfigFromURL(url string) (*rest.Config, error) {
	var data []byte
	if isValidUrl(url) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("url [%s] not found", url)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("url [%s] failed to read", url)
		}
	} else {
		exists, _ := utilfile.Exist(url)
		if !exists {
			return nil, fmt.Errorf("file [%s] not found", url)
		}
		var err error
		data, err = os.ReadFile(url)
		if err != nil {
			return nil, err
		}
	}

	var kubeconfig *clientcmdv1.Config
	err := yaml.Unmarshal(data, &kubeconfig)
	if err != nil {
		return nil, err
	}
	return RestConfigFromV1Config(kubeconfig)
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// RestConfigFromV1Config takes a v1 config and returns a kubeconfig
func RestConfigFromV1Config(kc *clientcmdv1.Config) (*rest.Config, error) {
	var c clientcmdapi.Config
	err := latest.Scheme.Convert(kc, &c, nil)
	if err != nil {
		return nil, err
	}

	kubeconfig := clientcmd.NewDefaultClientConfig(c, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

func MakeKubeconfig(server, namespace, token string) ([]byte, error) {
	return json.MarshalIndent(&clientcmdv1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []clientcmdv1.NamedCluster{
			{
				Name: "cluster",
				Cluster: clientcmdv1.Cluster{
					Server:                server,
					InsecureSkipTLSVerify: true,
				},
			},
		},
		AuthInfos: []clientcmdv1.NamedAuthInfo{
			{
				Name: "user",
				AuthInfo: clientcmdv1.AuthInfo{
					Token: token,
				},
			},
		},
		Contexts: []clientcmdv1.NamedContext{
			{
				Name: "cluster",
				Context: clientcmdv1.Context{
					Cluster:   "cluster",
					Namespace: namespace,
					AuthInfo:  "user",
				},
			},
		},
		CurrentContext: "cluster",
	}, "", "    ")
}
