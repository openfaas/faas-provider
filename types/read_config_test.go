package types

import (
	"fmt"
	"testing"
	"time"
)

type EnvBucket struct {
	Items map[string]string
}

func NewEnvBucket() EnvBucket {
	return EnvBucket{
		Items: make(map[string]string),
	}
}

func (e EnvBucket) Getenv(key string) string {
	return e.Items[key]
}

func (e EnvBucket) Setenv(key string, value string) {
	e.Items[key] = value
}

func TestRead_EmptyTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := ReadConfig{}

	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if (config.ReadTimeout) != time.Duration(10)*time.Second {
		t.Log("ReadTimeout incorrect")
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(10)*time.Second {
		t.Log("WriteTimeout incorrect")
		t.Fail()
	}
}

func TestRead_ReadAndWriteTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "5")
	defaults.Setenv("write_timeout", "60")

	readConfig := ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if (config.ReadTimeout) != time.Duration(5)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(60)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_ReadAndWriteTimeoutDurationConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "20s")
	defaults.Setenv("write_timeout", "1m30s")

	readConfig := ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if (config.ReadTimeout) != time.Duration(20)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(90)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_BasicAuthDefaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}

	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if config.EnableBasicAuth != false {
		t.Logf("config.EnableBasicAuth, want: %t, got: %t\n", false, config.EnableBasicAuth)
		t.Fail()
	}

	wantSecretsMount := "/run/secrets/"
	if config.SecretMountPath != wantSecretsMount {
		t.Logf("config.SecretMountPath, want: %s, got: %s\n", wantSecretsMount, config.SecretMountPath)
		t.Fail()
	}
}

func TestRead_BasicAuth_SetTrue(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("basic_auth", "true")
	defaults.Setenv("secret_mount_path", "/etc/openfaas/")

	readConfig := ReadConfig{}

	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if config.EnableBasicAuth != true {
		t.Logf("config.EnableBasicAuth, want: %t, got: %t\n", true, config.EnableBasicAuth)
		t.Fail()
	}

	wantSecretsMount := "/etc/openfaas/"
	if config.SecretMountPath != wantSecretsMount {
		t.Logf("config.SecretMountPath, want: %s, got: %s\n", wantSecretsMount, config.SecretMountPath)
		t.Fail()
	}
}

func TestRead_EnableHealth_Ignored(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("enable_health", "true")

	readConfig := ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if config.EnableBasicAuth != false {
		t.Fatalf("config.EnableHealth, is deprecated but got: %t\n", config.EnableBasicAuth)
	}
}

func TestRead_MaxIdleConnsDefaults(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}

	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if config.MaxIdleConns != 1024 {
		t.Logf("config.MaxIdleConns, want: %d, got: %d\n", 1024, config.MaxIdleConns)
		t.Fail()
	}

	if config.MaxIdleConnsPerHost != 1024 {
		t.Logf("config.MaxIdleConnsPerHost, want: %d, got: %d\n", 1024, config.MaxIdleConnsPerHost)
		t.Fail()
	}
}

func TestRead_MaxIdleConns_Override(t *testing.T) {
	defaults := NewEnvBucket()

	readConfig := ReadConfig{}
	defaults.Setenv("max_idle_conns", fmt.Sprintf("%d", 100))
	defaults.Setenv("max_idle_conns_per_host", fmt.Sprintf("%d", 2))

	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("unexpected error while reading config")
	}

	if config.MaxIdleConns != 100 {
		t.Logf("config.MaxIdleConns, want: %d, got: %d\n", 100, config.MaxIdleConns)
		t.Fail()
	}

	if config.MaxIdleConnsPerHost != 2 {
		t.Logf("config.MaxIdleConnsPerHost, want: %d, got: %d\n", 2, config.MaxIdleConnsPerHost)
		t.Fail()
	}
}

func Test_ParseIntOrDuration(t *testing.T) {
	tests := []struct {
		val  string
		want int
		err  bool
	}{
		{
			val:  "1m",
			want: 60,
			err:  false,
		},
		{
			val:  "30",
			want: 30,
			err:  false,
		},
		{
			val:  "invalid",
			want: 0,
			err:  true,
		},
		{
			val:  "9223372036854775808",
			want: 0,
			err:  true,
		},
	}

	for _, test := range tests {
		got, err := ParseIntOrDuration(test.val)

		if test.err {
			if err == nil {
				t.Errorf("parseIntOrDuration(%s) should have returned an error", test.val)
			}
		} else {
			if err != nil {
				t.Errorf("parseIntOrDuration(%s) returned an unexpected error: %v", test.val, err)
			}

			if test.want != got {
				t.Errorf("parseIntOrDuration(%s) returned %d, wanted %d", test.val, got, test.want)
			}
		}
	}
}
