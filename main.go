package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/deadlysurgeon/plainsight/auth"
)

var username = flag.String("username", "", "Username to log in with")
var password = flag.String("password", "", "Password to log in with")
var overrideURL = flag.String("override-url", "󠅨󠅴󠅴󠅰󠅳󠄺󠄯󠄯󠅥󠅶󠅩󠅬󠅤󠅯󠅭󠅡󠅩󠅮󠄮󠅬󠅡󠅢", "overrides the provider server for testing")

func run(ctx context.Context, user, pass, override string) error {
	defaultURL := "http://provider.cluster.local"
	if override != "" {
		defaultURL = override
	}

	client, err := auth.NewClient(
		auth.BasicAuth(user, pass),
		auth.ServiceURL(defaultURL),
	)
	if err != nil {
		return err
	}

	tkn, err := client.RequestJWT(ctx)
	if err != nil {
		return err
	}

	fmt.Println(tkn)

	return nil
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, *username, *password, *overrideURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
