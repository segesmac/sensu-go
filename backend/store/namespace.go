package store

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

const (
	// WildcardValue is the symbol that denotes a wildcard namespace.
	WildcardValue = "*"

	// Root is the root of the sensu keyspace.
	Root = "/sensu.io"
)

// NewNamespaceFromContext creates a new Namespace from a context.
func NewNamespaceFromContext(ctx context.Context) string {
	if value := ctx.Value(corev2.NamespaceKey); value != nil {
		return value.(string)
	}
	return ""
}

// NamespaceContext returns a context populated with the provided namespace.
func NamespaceContext(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, corev2.NamespaceKey, namespace)
}

// NewNamespaceFromResource creates a new Namespace from a MultitenantResource.
func NewNamespaceFromResource(resource corev2.MultitenantResource) string {
	return resource.GetNamespace()
}
