package fakes

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

type Lister struct {
	NamespaceLister *NamespaceLister
}

func NewLister() *Lister {
	return &Lister{
		NamespaceLister: NewNamespaceLister(),
	}
}

func (l *Lister) List(selector labels.Selector) (ret []*v1.Secret, err error) {
	fmt.Printf("Lister List: %s\n", selector.String())
	return []*v1.Secret{}, errors.NewNotFound(schema.GroupResource{}, selector.String())
}

func (l *Lister) Secrets(namespace string) listersv1.SecretNamespaceLister {
	return l.NamespaceLister
}

func (l *Lister) Set(name string, secret *v1.Secret) {
	l.NamespaceLister.Set(name, secret)
}
