package kube

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNameSpace creates a new namespace which will contain all the dependencies of a Course.
func CreateNameSpace(courseName string) error {
	clientset, err := getClient()
	if err != nil {
		return err
	}
	ns := &apiv1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NameSpace",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: courseName,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(ns)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNameSpace deletes a namespace that represents a specific Course.
// Warning: This method will delete all the Course's objects and its dependencies from the system.
func DeleteNameSpace(courseName string) error {
	clientset, err := getClient()
	if err != nil {
		return err
	}
	err = clientset.CoreV1().Namespaces().Delete(courseName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// AllNamespaces lists all the namespaces in the system.
func AllNamespaces() (string, error) {
	clientset, err := getClient()
	if err != nil {
		return "", err
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	return namespaces.String(), nil
}
