package test

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type metaContext string

const (
	correlationID metaContext = "correlation-id"
	responseHDR   metaContext = "my-response-header"
	responseTRLR  metaContext = "correlation-id-consumed"
)

func clientBefore(ctx context.Context, md *metadata.MD) context.Context {
	if hdr, ok := ctx.Value(correlationID).(string); ok {
		(*md)[string(correlationID)] = append((*md)[string(correlationID)], hdr)
	}
	if len(*md) > 0 {
		fmt.Println("\tClient >> Request Headers:")
		for key, val := range *md {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}
	return ctx
}

func serverBefore(ctx context.Context, md *metadata.MD) context.Context {
	if len(*md) > 0 {
		fmt.Println("\tServer << Request Headers:")
		for key, val := range *md {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}
	if hdr, ok := (*md)[string(correlationID)]; ok {
		cID := hdr[len(hdr)-1]
		ctx = context.WithValue(ctx, correlationID, cID)
		fmt.Printf("\tServer placed correlationID %q in context\n", cID)
	}
	return ctx
}

func serverAfter(ctx context.Context, _ *metadata.MD) {
	var mdHeader, mdTrailer metadata.MD

	mdHeader = metadata.Pairs(string(responseHDR), "has-a-value")
	if err := grpc.SendHeader(ctx, mdHeader); err != nil {
		log.Fatalf("unable to send header: %+v\n", err)
	}

	if hdr, ok := ctx.Value(correlationID).(string); ok {
		mdTrailer = metadata.Pairs(string(responseTRLR), hdr)
		if err := grpc.SetTrailer(ctx, mdTrailer); err != nil {
			log.Fatalf("unable to set trailer: %+v\n", err)
		}
		fmt.Printf("\tServer found correlationID %q in context, set consumed trailer\n", hdr)
	}
	if len(mdHeader) > 0 {
		fmt.Println("\tServer >> Response Headers:")
		for key, val := range mdHeader {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}
	if len(mdTrailer) > 0 {
		fmt.Println("\tServer >> Response Trailers:")
		for key, val := range mdTrailer {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}
}

func clientAfter(ctx context.Context, mdHeader metadata.MD, mdTrailer metadata.MD) context.Context {
	if len(mdHeader) > 0 {
		fmt.Println("\tClient << Response Headers:")
		for key, val := range mdHeader {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}
	if len(mdTrailer) > 0 {
		fmt.Println("\tClient << Response Trailers:")
		for key, val := range mdTrailer {
			fmt.Printf("\t\t%s: %s\n", key, val[len(val)-1])
		}
	}

	if hdr, ok := mdTrailer[string(responseTRLR)]; ok {
		ctx = context.WithValue(ctx, responseTRLR, hdr[len(hdr)-1])
	}
	return ctx
}

func SetCorrelationID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, correlationID, v)
}

func GetConsumedCorrelationID(ctx context.Context) string {
	if trlr, ok := ctx.Value(responseTRLR).(string); ok {
		return trlr
	}
	return ""
}
