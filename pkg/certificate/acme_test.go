package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/appscode/voyager/api"
	"github.com/appscode/voyager/test/testframework"
	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/acme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func TestNewDomainCollection(t *testing.T) {
	d := NewDomainCollection("a.com")
	assert.Equal(t, `["a.com"]`, d.String())

	d.Append("hello.world").Append("foo.bar")
	assert.Equal(t, `["a.com","hello.world","foo.bar"]`, d.String())
}

func TestACMECertData(t *testing.T) {
	certificateSecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultCertPrefix + "hello",
			Namespace: "default",
			Labels: map[string]string{
				certificateKey:              "true",
				certificateKey + "/domains": NewDomainCollection("appscode.com").String(),
			},
			Annotations: map[string]string{
				certificateKey: "true",
			},
		},
		Data: map[string][]byte{
			apiv1.TLSCertKey:       []byte("Certificate key"),
			apiv1.TLSPrivateKeyKey: []byte("Certificate private key"),
		},
		Type: apiv1.SecretTypeTLS,
	}

	cert, err := NewACMECertDataFromSecret(certificateSecret, &api.Certificate{})
	assert.Nil(t, err)

	convertedCert := cert.ToSecret("hello", "default")
	assert.Equal(t, certificateSecret, convertedCert)
}

func TestACMECertDataError(t *testing.T) {
	certificateSecret := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultCertPrefix + "hello",
			Namespace: "default",
			Labels: map[string]string{
				certificateKey:              "true",
				certificateKey + "/domains": NewDomainCollection("appscode.com").String(),
			},
			Annotations: map[string]string{
				certificateKey: "true",
			},
		},
		Data: map[string][]byte{
			apiv1.TLSPrivateKeyKey: []byte("Certificate private key"),
		},
		Type: apiv1.SecretTypeTLS,
	}

	_, err := NewACMECertDataFromSecret(certificateSecret, &api.Certificate{})
	assert.NotNil(t, err)
	assert.Equal(t, "INTERNAL:Could not find key tls.crt in secret "+defaultCertPrefix+"hello", err.Error())

}

func TestClient(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := &ACMEUserData{
		Email:        "test@test.com",
		Registration: new(acme.RegistrationResource),
		Key: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
	}

	config := &ACMEConfig{
		Provider: "http",
		UserData: user,
	}
	_, err = NewACMEClient(config)
	assert.Nil(t, err)

	if testframework.TestContext.Verbose {
		config := &ACMEConfig{
			Provider: "http",
			UserData: user,
			ProviderCredentials: map[string][]byte{
				"GCE_SERVICE_ACCOUNT_DATA": []byte(os.Getenv("TEST_GCE_SERVICE_ACCOUNT_DATA")),
				"GCE_PROJECT":              []byte(os.Getenv("TEST_GCE_PROJECT")),
			},
		}
		_, err = NewACMEClient(config)
		assert.Nil(t, err)
	}
}