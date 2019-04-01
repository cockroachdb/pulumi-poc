package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/backend"
	"github.com/pulumi/pulumi/pkg/backend/display"
	"github.com/pulumi/pulumi/pkg/backend/httpstate"
	"github.com/pulumi/pulumi/pkg/diag/colors"
	"github.com/pulumi/pulumi/pkg/engine"
	"github.com/pulumi/pulumi/pkg/util/cancel"
	"github.com/pulumi/pulumi/pkg/util/cmdutil"
	"github.com/pulumi/pulumi/pkg/util/result"
	"github.com/pulumi/pulumi/pkg/workspace"
)

// Copied from Pulumi souce code and modified.
func loginAndGetBackend(opts display.Options) (backend.Backend, error) {
	// I guess eventually these creds would have to be stored in vault??
	creds, err := workspace.GetStoredCredentials()
	if err != nil {
		return nil, err
	}
	return httpstate.Login(context.TODO(), cmdutil.Diag(), creds.Current, "", opts)
}

// Copied from Pulumi souce code and modified.
func loginAndGetStack(
	stackName string, opts display.Options) (backend.Stack, error) {
	b, err := loginAndGetBackend(opts)
	if err != nil {
		return nil, err
	}

	stackRef, err := b.ParseStackReference(stackName)
	if err != nil {
		return nil, err
	}

	stack, err := b.GetStack(context.TODO(), stackRef)
	if err != nil {
		return nil, err
	}
	if stack != nil {
		return stack, err
	}

	return nil, errors.Errorf("no stack named '%s' found", stackName)
}

// Copied from Pulumi source code, since not all parts are exported. Maybe there
// is a better way...
type cancellationScope struct {
	context *cancel.Context
	sigint  chan os.Signal
	done    chan bool
}

func (s *cancellationScope) Context() *cancel.Context {
	return s.context
}

func (s *cancellationScope) Close() {
	signal.Stop(s.sigint)
	close(s.sigint)
	<-s.done
}

type cancellationScopeSource int

var cancellationScopes = backend.CancellationScopeSource(cancellationScopeSource(0))

func (cancellationScopeSource) NewScope(events chan<- engine.Event, isPreview bool) backend.CancellationScope {
	cancelContext, cancelSource := cancel.NewContext(context.Background())

	c := &cancellationScope{
		context: cancelContext,
		sigint:  make(chan os.Signal),
		done:    make(chan bool),
	}

	go func() {
		for range c.sigint {
			// If we haven't yet received a SIGINT, call the cancellation func. Otherwise call the termination
			// func.
			if cancelContext.CancelErr() == nil {
				message := "^C received; cancelling. If you would like to terminate immediately, press ^C again.\n"
				if !isPreview {
					message += colors.BrightRed + "Note that terminating immediately may lead to orphaned resources " +
						"and other inconsistencies.\n" + colors.Reset
				}
				events <- engine.Event{
					Type: engine.StdoutColorEvent,
					Payload: engine.StdoutEventPayload{
						Message: message,
						Color:   colors.Always,
					},
				}

				cancelSource.Cancel()
			} else {
				message := colors.BrightRed + "^C received; terminating" + colors.Reset
				events <- engine.Event{
					Type: engine.StdoutColorEvent,
					Payload: engine.StdoutEventPayload{
						Message: message,
						Color:   colors.Always,
					},
				}

				cancelSource.Terminate()
			}
		}
		close(c.done)
	}()
	signal.Notify(c.sigint, os.Interrupt)

	return c
}

func DiffOrUpdate(dryrun bool) (string, error) {
	opts := backend.UpdateOptions{
		Engine: engine.UpdateOptions{},
		AutoApprove: true,
		Display: display.Options{
			Color: colors.Always,
		},
	}

	// Login to Pulumi service and get stack. Eventually there would be 1+
	// stack(s) per customer cluster.
	s, err := loginAndGetStack("joshimhoff/poc/dev", opts.Display)
	if err != nil {
		return "", err
	}
	p := &workspace.Project{
		Name: "poc",
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{}),
	}

	// What is returned from this API is NOT a detailed plan, despite its name.
	// As a side effect of calling s.Preview, a detailed plan is logged to
	// stdout. Not sure what best way of getting at the plan is yet.
	var changes engine.ResourceChanges
	var res result.Result
	if dryrun {
		changes, res = s.Preview(context.TODO(), backend.UpdateOperation{
			Proj:   p,
			Root:   pulumiDir,
			M:      &backend.UpdateMetadata{},
			Opts:   opts,
			Scopes: cancellationScopes,
		})
	} else {
		changes, res = s.Update(context.TODO(), backend.UpdateOperation{
			Proj:   p,
			Root:   pulumiDir,
			M:      &backend.UpdateMetadata{},
			Opts:   opts,
			Scopes: cancellationScopes,
		})
	}
	if res != nil {
		return "", res.Error()
	}
	return fmt.Sprintf("%+v", changes), nil
}
