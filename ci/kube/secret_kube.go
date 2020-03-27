package kube

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Jobsecrets stores the data that is passed in a Secret object.
func Jobsecrets(id string, namespace string, pass string) error {
	clientset, err := getClient()
	if err != nil {
		return err
	}

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

	_, err = clientset.CoreV1().Secrets(namespace).Create(newSec)
	if err != nil {
		return err
	}
	return nil
}

//DeleteJobSecret delete a specified Secret object.
func DeleteJobSecret(id string, namespace string) error {
	clientset, err := getClient()
	if err != nil {
		return err
	}
	err = clientset.CoreV1().Secrets(namespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//GetJobSecret returns the data that is passed to a specified Secret object.
func GetJobSecret(id string, namespace string) (string, error) {
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
