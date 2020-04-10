package kube

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// makeSecret creates a Secret Object.
// The parameter "pass" is data that needs to be stored secretly.
func makeSecret(clientset *kubernetes.Clientset, id string, namespace string, pass string) error {

	newSec := &apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			id: []byte(pass),
		},
		Type: "Opaque",
	}

	_, err := clientset.CoreV1().Secrets(namespace).Create(newSec)
	if err != nil {
		return err
	}
	return nil
}

// removeSecret delete the Secret object.
func removeSecret(clientset *kubernetes.Clientset, id string, namespace string) error {
	err := clientset.CoreV1().Secrets(namespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// getSecretData returns secret data of the Secret object.
func getSecretData(id string, namespace string) (string, error) {
	clientset, err := getClient()
	if err != nil {
		return "", err
	}
	sec, err := clientset.CoreV1().Secrets(namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	extractSecret := sec.Data[id]
	return string(extractSecret), nil
}
