package logging

import (
	"context"
	"sort"

	"github.com/muhlemmer/gu"
	"golang.org/x/exp/slog"
)

const (
	ContextDataIDKey = "id"
)

type ContextData map[string]any

func NewContextData(id string) ContextData {
	data := make(ContextData)
	if id != "" {
		data[ContextDataIDKey] = id
	}
	return data
}

// LogValue implements [slog.LogValuer].
//
// EXPERIMENTAL: API will break when we switch from `x/exp/slog` to `log/slog`
// when we drop Go <1.21 support.
func (d ContextData) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(d))
	for k, v := range d {
		attrs = append(attrs, slog.Any(k, v))
	}

	// TODO: switch to slices.SortFunc after
	// <1.21 support drop.
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].Key < attrs[j].Key
	})
	return slog.GroupValue(attrs...)
}

func (d ContextData) Clone() ContextData {
	return gu.MapCopy(d)
}

type requestDataKeyType struct{}

var requestDataKey requestDataKeyType

func DataFromContext(ctx context.Context) (ContextData, bool) {
	data, ok := ctx.Value(requestDataKey).(ContextData)
	return data, ok
}

// ContextWithData adds data to the context.
// The data is merged with any previsously set data.
// Keys in the existing data are overwritten with
// the passed data when duplicates exist.
//
// Warning: The data is transformed into Attributes when
// the context is passed to the logger.
// This means the attributes will be part of
// the Group the Logger is currently in.
func ContextWithData(ctx context.Context, data ContextData) context.Context {
	out, ok := DataFromContext(ctx)
	if !ok {
		return context.WithValue(ctx, requestDataKey, data)
	}
	out = out.Clone()
	gu.MapMerge(data, out)
	return context.WithValue(ctx, requestDataKey, out)
}
